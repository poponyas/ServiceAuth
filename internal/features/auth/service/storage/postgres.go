package service_storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poponyas/AuthService/internal/core/domain"
)

type Storage struct{ DB *pgxpool.Pool }

func (s *Storage) SaveUser(ctx context.Context, email string, hash []byte) (int64, error) {
	var id int64
	err := s.DB.QueryRow(ctx, "INSERT INTO users(email,pass_hash) VALUES($1,$2) RETURNING id", email, hash).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, ErrUserExists
	}
	return id, err
}
func (s *Storage) User(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	err := s.DB.QueryRow(ctx, "SELECT id,email,pass_hash FROM users WHERE email=$1", email).Scan(&u.ID, &u.Email, &u.PassHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrUserNotFound
	}
	return u, err
}
func (s *Storage) IsAdmin(ctx context.Context, id int64) (bool, error) {
	var admin bool
	err := s.DB.QueryRow(ctx, "SELECT is_admin FROM users WHERE id=$1", id).Scan(&admin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrUserNotFound
	}
	return admin, err
}
func (s *Storage) App(ctx context.Context, id int) (domain.App, error) {
	var app domain.App
	err := s.DB.QueryRow(ctx, "SELECT id,name FROM apps WHERE id=$1", id).Scan(&app.ID, &app.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return app, ErrAppNotFound
	}
	if err != nil {
		return app, fmt.Errorf("app query: %w", err)
	}
	return app, nil
}
