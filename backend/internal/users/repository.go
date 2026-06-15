package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"afs-todo-backend/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var exists bool
	err := r.db.QueryRowContext(
		queryCtx,
		"SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)",
		normalizeEmail(email),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check user email exists: %w", err)
	}

	return exists, nil
}

func (r *Repository) Create(ctx context.Context, name string, email string, passwordHash string) (models.User, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user := models.User{}
	err := r.db.QueryRowContext(queryCtx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at, updated_at
	`, name, normalizeEmail(email), passwordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrEmailAlreadyExists
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (models.UserWithPassword, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user := models.UserWithPassword{}
	err := r.db.QueryRowContext(queryCtx, `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`, normalizeEmail(email)).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserWithPassword{}, ErrUserNotFound
		}
		return models.UserWithPassword{}, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (models.User, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user := models.User{}
	err := r.db.QueryRowContext(queryCtx, `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
