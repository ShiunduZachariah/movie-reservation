package service

import (
	"context"
	"time"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type ShowtimeService struct {
	db        *pgxpool.Pool
	showtimes *repository.ShowtimeRepository
	screens   *repository.ScreenRepository
	seats     *repository.SeatRepository
}

type CreateShowtimeInput struct {
	MovieID     uuid.UUID
	ScreenID    uuid.UUID
	StartTime   time.Time
	EndTime     time.Time
	TicketPrice decimal.Decimal
}

func NewShowtimeService(db *pgxpool.Pool, showtimes *repository.ShowtimeRepository, screens *repository.ScreenRepository, seats *repository.SeatRepository) *ShowtimeService {
	return &ShowtimeService{db: db, showtimes: showtimes, screens: screens, seats: seats}
}

func (s *ShowtimeService) Create(ctx context.Context, input CreateShowtimeInput) (*model.Showtime, error) {
	screen, err := s.screens.GetByID(ctx, input.ScreenID)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	showtime, err := s.showtimes.Create(ctx, tx, &model.Showtime{
		MovieID:        input.MovieID,
		ScreenID:       input.ScreenID,
		StartTime:      input.StartTime,
		EndTime:        input.EndTime,
		TicketPrice:    input.TicketPrice,
		AvailableSeats: screen.TotalSeats,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.showtimes.GetByID(ctx, showtime.ID)
}

func (s *ShowtimeService) ListByDate(ctx context.Context, date string) ([]*model.Showtime, error) {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errs.BadRequest("INVALID_DATE", "date must use YYYY-MM-DD", nil)
	}
	return s.showtimes.ListByDate(ctx, parsed)
}

func (s *ShowtimeService) ListByMovie(ctx context.Context, movieID string) ([]*model.Showtime, error) {
	id, err := uuid.Parse(movieID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_MOVIE_ID", "invalid movie id", nil)
	}
	return s.showtimes.ListByMovie(ctx, id)
}

func (s *ShowtimeService) AvailableSeats(ctx context.Context, showtimeID string) ([]*model.Seat, error) {
	id, err := uuid.Parse(showtimeID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_SHOWTIME_ID", "invalid showtime id", nil)
	}
	return s.seats.ListAvailableByShowtime(ctx, id)
}
