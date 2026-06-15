package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
)

//go:embed *.sql
var migrationFiles embed.FS

func Run(ctx context.Context, db *sql.DB) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	files, err := fs.Glob(migrationFiles, "*.sql")
	if err != nil {
		return fmt.Errorf("list migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		if err := runMigration(ctx, db, file); err != nil {
			return err
		}
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	return nil
}

func runMigration(ctx context.Context, db *sql.DB, file string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", file, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var alreadyApplied bool
	if err = tx.QueryRowContext(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)",
		file,
	).Scan(&alreadyApplied); err != nil {
		return fmt.Errorf("check migration %s: %w", file, err)
	}

	if alreadyApplied {
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit skipped migration %s: %w", file, err)
		}
		return nil
	}

	sqlBytes, err := migrationFiles.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", file, err)
	}

	if _, err = tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("execute migration %s: %w", file, err)
	}

	if _, err = tx.ExecContext(
		ctx,
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		file,
	); err != nil {
		return fmt.Errorf("record migration %s: %w", file, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", file, err)
	}

	log.Printf("applied migration %s", file)
	return nil
}
