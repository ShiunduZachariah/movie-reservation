package service

import (
	"context"
	"time"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminService struct {
	db    *pgxpool.Pool
	repos *repository.Repositories
	users *UserService
}

func NewAdminService(db *pgxpool.Pool, repos *repository.Repositories, users *UserService) *AdminService {
	return &AdminService{db: db, repos: repos, users: users}
}

func (s *AdminService) ListReservations(ctx context.Context) ([]*model.Reservation, error) {
	return s.repos.Reservations.ListAll(ctx)
}

func (s *AdminService) RevenueSummary(ctx context.Context, from, to string) ([]repository.RevenueRow, string, error) {
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, "", errs.BadRequest("INVALID_FROM_DATE", "from must use YYYY-MM-DD", nil)
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, "", errs.BadRequest("INVALID_TO_DATE", "to must use YYYY-MM-DD", nil)
	}
	rows, total, err := s.repos.Reservations.RevenueSummary(ctx, fromDate, toDate.Add(24*time.Hour))
	if err != nil {
		return nil, "", err
	}
	return rows, total.StringFixed(2), nil
}

func (s *AdminService) CapacitySummary(ctx context.Context) ([]repository.CapacityRow, error) {
	return s.repos.Reservations.CapacitySummary(ctx)
}
