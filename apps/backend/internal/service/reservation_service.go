package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/job"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/utils"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type ReservationService struct {
	cfg         *config.Config
	db          *pgxpool.Pool
	repos       *repository.Repositories
	jobs        *job.Service
	queueClient *azqueue.QueueClient
	logger      *zerolog.Logger
}

type CreateReservationInput struct {
	UserID     uuid.UUID
	ShowtimeID uuid.UUID
	SeatIDs    []uuid.UUID
}

func NewReservationService(cfg *config.Config, db *pgxpool.Pool, repos *repository.Repositories, jobs *job.Service, queueClient *azqueue.QueueClient, logger *zerolog.Logger) *ReservationService {
	return &ReservationService{cfg: cfg, db: db, repos: repos, jobs: jobs, queueClient: queueClient, logger: logger}
}

func (s *ReservationService) ReserveSeats(ctx context.Context, input CreateReservationInput) (*model.Reservation, error) {
	if len(input.SeatIDs) == 0 {
		return nil, errs.BadRequest("SEATS_REQUIRED", "at least one seat is required", nil)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	showtime, err := s.repos.Showtimes.GetForUpdate(ctx, tx, input.ShowtimeID)
	if err != nil {
		return nil, err
	}

	seats, err := s.repos.Seats.GetByIDs(ctx, input.SeatIDs)
	if err != nil {
		return nil, err
	}
	if len(seats) != len(input.SeatIDs) {
		return nil, errs.BadRequest("INVALID_SEATS", "one or more seats were not found", nil)
	}
	for _, seat := range seats {
		if seat.ScreenID != showtime.ScreenID {
			return nil, errs.BadRequest("SEAT_SCREEN_MISMATCH", "all seats must belong to the showtime screen", nil)
		}
	}

	takenSeats, err := s.repos.Seats.ListTakenByShowtime(ctx, input.ShowtimeID, input.SeatIDs)
	if err != nil {
		return nil, err
	}
	if len(takenSeats) > 0 {
		return nil, errs.Conflict("SEATS_NOT_AVAILABLE", "one or more seats are already reserved")
	}

	totalPrice := showtime.TicketPrice.Mul(decimal.NewFromInt(int64(len(input.SeatIDs))))
	reservation, err := s.repos.Reservations.Create(ctx, tx, &model.Reservation{
		UserID:     input.UserID,
		ShowtimeID: input.ShowtimeID,
		Status:     model.ReservationConfirmed,
		TotalPrice: totalPrice,
		ExpiresAt:  showtime.StartTime,
	})
	if err != nil {
		return nil, err
	}

	if err := s.repos.Reservations.CreateSeats(ctx, tx, reservation.ID, input.SeatIDs); err != nil {
		return nil, err
	}
	if err := s.repos.Showtimes.DecrementAvailableSeats(ctx, tx, input.ShowtimeID, len(input.SeatIDs)); err != nil {
		return nil, errs.Conflict("SEATS_NOT_AVAILABLE", "not enough seats available")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	reservation, err = s.repos.Reservations.GetByID(ctx, reservation.ID)
	if err != nil {
		return nil, err
	}

	if enqueueErr := s.enqueueTicketConfirmation(ctx, reservation); enqueueErr != nil {
		s.logger.Error().Err(enqueueErr).Str("reservation_id", reservation.ID.String()).Msg("failed to enqueue ticket confirmation email")
	}

	return reservation, nil
}

func (s *ReservationService) ListMyReservations(ctx context.Context, userID string) ([]*model.Reservation, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_USER_ID", "invalid user id", nil)
	}
	return s.repos.Reservations.ListByUser(ctx, id)
}

func (s *ReservationService) GetReservation(ctx context.Context, reservationID string) (*model.Reservation, error) {
	id, err := uuid.Parse(reservationID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_RESERVATION_ID", "invalid reservation id", nil)
	}
	return s.repos.Reservations.GetByID(ctx, id)
}

func (s *ReservationService) CancelReservation(ctx context.Context, reservationID string, userID string) (*model.Reservation, error) {
	resID, err := uuid.Parse(reservationID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_RESERVATION_ID", "invalid reservation id", nil)
	}
	usrID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_USER_ID", "invalid user id", nil)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	reservation, err := s.repos.Reservations.Cancel(ctx, tx, resID, usrID)
	if err != nil {
		return nil, err
	}

	seatCount, err := s.repos.Reservations.GetSeatCountTx(ctx, tx, reservation.ID)
	if err != nil {
		return nil, err
	}
	if err := s.repos.Showtimes.IncrementAvailableSeats(ctx, tx, reservation.ShowtimeID, seatCount); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.repos.Reservations.GetByID(ctx, reservation.ID)
}

func (s *ReservationService) enqueueTicketConfirmation(ctx context.Context, reservation *model.Reservation) error {
	seatLabels := make([]string, 0, len(reservation.Seats))
	for _, seat := range reservation.Seats {
		seatLabels = append(seatLabels, utils.SeatLabel(seat.Row, seat.Number))
	}

	payload := job.TicketConfirmationPayload{
		UserEmail:      reservation.UserEmail,
		UserName:       reservation.UserName,
		ReservationID:  reservation.ID.String(),
		MovieTitle:     reservation.MovieTitle,
		ShowDate:       reservation.ShowStart.Format("Monday, 2 January 2006"),
		ShowTime:       reservation.ShowStart.Format("3:04 PM"),
		ScreenName:     reservation.ScreenName,
		Seats:          seatLabels,
		TotalPrice:     utils.FormatKES(reservation.TotalPrice),
		ReservationURL: fmt.Sprintf("%s/reservations/%s", s.cfg.App.BaseURL, reservation.ID),
	}

	if s.cfg.Primary.Env == "production" && s.queueClient != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		encoded := base64.StdEncoding.EncodeToString(body)
		_, err = s.queueClient.EnqueueMessage(ctx, encoded, nil)
		return err
	}

	if s.jobs == nil {
		return nil
	}
	task, err := job.NewTicketConfirmationTask(payload)
	if err != nil {
		return err
	}
	return s.jobs.Enqueue(ctx, task)
}
