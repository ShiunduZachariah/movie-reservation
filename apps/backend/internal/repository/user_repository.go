package repository

import (
	"context"
	"fmt"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/sqlerr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByClerkID(ctx context.Context, clerkID string) (*model.User, error) {
	const query = `
		SELECT id, clerk_id, email, name, role, password_hash, created_at, updated_at
		FROM users
		WHERE clerk_id = $1
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, clerkID).Scan(
		&user.ID,
		&user.ClerkID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("table:users:%w", err))
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	const query = `
		SELECT id, clerk_id, email, name, role, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.ClerkID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("table:users:%w", err))
	}

	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, db DBTX, clerkID, email, name string) (*model.User, error) {
	const query = `
		INSERT INTO users (clerk_id, email, name)
		VALUES ($1, $2, $3)
		RETURNING id, clerk_id, email, name, role, password_hash, created_at, updated_at
	`

	var user model.User
	if err := db.QueryRow(ctx, query, clerkID, email, name).Scan(
		&user.ID,
		&user.ClerkID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const query = `
		SELECT id, clerk_id, email, name, role, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.ClerkID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("table:users:%w", err))
	}

	return &user, nil
}

func (r *UserRepository) CreateLocal(ctx context.Context, db DBTX, email, name, password string, role model.UserRole) (*model.User, error) {
	const query = `
		INSERT INTO users (clerk_id, email, name, role, password_hash)
		VALUES ($1, $2, $3, $4, crypt($5, gen_salt('bf')))
		RETURNING id, clerk_id, email, name, role, password_hash, created_at, updated_at
	`

	clerkID := "local:" + email
	var user model.User
	if err := db.QueryRow(ctx, query, clerkID, email, name, role, password).Scan(
		&user.ID,
		&user.ClerkID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	return &user, nil
}

func (r *UserRepository) AuthenticateLocal(ctx context.Context, email, password string) (*model.User, error) {
	const query = `
		SELECT id, clerk_id, email, name, role, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
		  AND password_hash IS NOT NULL
		  AND password_hash = crypt($2, password_hash)
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, email, password).Scan(
		&user.ID,
		&user.ClerkID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("table:users:%w", err))
	}

	return &user, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, db DBTX, userID uuid.UUID, role model.UserRole) error {
	tag, err := db.Exec(ctx, `
		UPDATE users
		SET role = $2, updated_at = NOW()
		WHERE id = $1
	`, userID, role)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	if tag.RowsAffected() == 0 {
		return sqlerr.HandleError(pgx.ErrNoRows)
	}
	return nil
}
