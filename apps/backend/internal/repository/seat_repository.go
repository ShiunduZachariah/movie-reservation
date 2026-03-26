package repository

import (
	"context"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/sqlerr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeatRepository struct {
	pool *pgxpool.Pool
}

func NewSeatRepository(pool *pgxpool.Pool) *SeatRepository {
	return &SeatRepository{pool: pool}
}

func (r *SeatRepository) ListByScreen(ctx context.Context, screenID uuid.UUID) ([]*model.Seat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, screen_id, row, number, created_at, updated_at
		FROM seats
		WHERE screen_id = $1
		ORDER BY row ASC, number ASC
	`, screenID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var seats []*model.Seat
	for rows.Next() {
		var seat model.Seat
		if err := rows.Scan(&seat.ID, &seat.ScreenID, &seat.Row, &seat.Number, &seat.CreatedAt, &seat.UpdatedAt); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		seats = append(seats, &seat)
	}
	return seats, nil
}

func (r *SeatRepository) ListAvailableByShowtime(ctx context.Context, showtimeID uuid.UUID) ([]*model.Seat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.screen_id, s.row, s.number, s.created_at, s.updated_at
		FROM seats s
		WHERE s.screen_id = (SELECT screen_id FROM showtimes WHERE id = $1)
		  AND s.id NOT IN (
			  SELECT rs.seat_id
			  FROM reservation_seats rs
			  JOIN reservations r ON r.id = rs.reservation_id
			  WHERE r.showtime_id = $1
			    AND r.status IN ('pending', 'confirmed')
		  )
		ORDER BY s.row ASC, s.number ASC
	`, showtimeID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var seats []*model.Seat
	for rows.Next() {
		var seat model.Seat
		if err := rows.Scan(&seat.ID, &seat.ScreenID, &seat.Row, &seat.Number, &seat.CreatedAt, &seat.UpdatedAt); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		seats = append(seats, &seat)
	}
	return seats, nil
}

func (r *SeatRepository) ListTakenByShowtime(ctx context.Context, showtimeID uuid.UUID, seatIDs []uuid.UUID) ([]*model.Seat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.screen_id, s.row, s.number, s.created_at, s.updated_at
		FROM seats s
		JOIN reservation_seats rs ON rs.seat_id = s.id
		JOIN reservations r ON r.id = rs.reservation_id
		WHERE r.showtime_id = $1
		  AND r.status IN ('pending', 'confirmed')
		  AND s.id = ANY($2)
	`, showtimeID, seatIDs)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var seats []*model.Seat
	for rows.Next() {
		var seat model.Seat
		if err := rows.Scan(&seat.ID, &seat.ScreenID, &seat.Row, &seat.Number, &seat.CreatedAt, &seat.UpdatedAt); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		seats = append(seats, &seat)
	}
	return seats, nil
}

func (r *SeatRepository) GetByIDs(ctx context.Context, seatIDs []uuid.UUID) ([]*model.Seat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, screen_id, row, number, created_at, updated_at
		FROM seats
		WHERE id = ANY($1)
		ORDER BY row ASC, number ASC
	`, seatIDs)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var seats []*model.Seat
	for rows.Next() {
		var seat model.Seat
		if err := rows.Scan(&seat.ID, &seat.ScreenID, &seat.Row, &seat.Number, &seat.CreatedAt, &seat.UpdatedAt); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		seats = append(seats, &seat)
	}
	return seats, nil
}
