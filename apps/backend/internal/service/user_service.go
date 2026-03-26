package service

import (
	"context"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	db    *pgxpool.Pool
	users *repository.UserRepository
}

func NewUserService(db *pgxpool.Pool, users *repository.UserRepository) *UserService {
	return &UserService{db: db, users: users}
}

func (s *UserService) SyncUser(ctx context.Context, clerkID, email, name string) (*model.User, error) {
	user, err := s.users.GetByClerkID(ctx, clerkID)
	if err == nil {
		return user, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	user, err = s.users.Create(ctx, tx, clerkID, email, name)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Promote(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errs.BadRequest("INVALID_USER_ID", "invalid user id", nil)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.users.UpdateRole(ctx, tx, id, model.RoleAdmin); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
