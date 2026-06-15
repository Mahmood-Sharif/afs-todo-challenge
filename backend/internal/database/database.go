package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"

	"afs-todo-backend/internal/config"
	"afs-todo-backend/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database struct {
	conn *sql.DB
}

func Connect(ctx context.Context, cfg config.DatabaseConfig) (*Database, error) {
	db, err := sql.Open("pgx", connectionString(cfg))
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := migrations.Run(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &Database{conn: db}, nil
}

func (db *Database) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.conn.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) SchemaTables(ctx context.Context) ([]string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := db.conn.QueryContext(queryCtx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
			AND table_name IN ('users', 'todos')
	`)
	if err != nil {
		return nil, fmt.Errorf("query schema tables: %w", err)
	}
	defer rows.Close()

	found := map[string]bool{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("scan schema table: %w", err)
		}
		found[tableName] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schema tables: %w", err)
	}

	tables := make([]string, 0, 2)
	for _, tableName := range []string{"users", "todos"} {
		if found[tableName] {
			tables = append(tables, tableName)
		}
	}

	return tables, nil
}

func connectionString(cfg config.DatabaseConfig) string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}

	query := dsn.Query()
	query.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
}
