package repository

import (
	"context"
	"time"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/sqlerr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type ReservationRepository struct {
	pool *pgxpool.Pool
}

func NewReservationRepository(pool *pgxpool.Pool) *ReservationRepository {
	return &ReservationRepository{pool: pool}
}

func (r *ReservationRepository) Create(ctx context.Context, tx pgx.Tx, reservation *model.Reservation) (*model.Reservation, error) {
	var created model.Reservation
	if err := tx.QueryRow(ctx, `
		INSERT INTO reservations (user_id, showtime_id, status, total_price, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, showtime_id, status, total_price, expires_at, created_at, updated_at
	`, reservation.UserID, reservation.ShowtimeID, reservation.Status, reservation.TotalPrice, reservation.ExpiresAt).Scan(
		&created.ID,
		&created.UserID,
		&created.ShowtimeID,
		&created.Status,
		&created.TotalPrice,
		&created.ExpiresAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &created, nil
}

func (r *ReservationRepository) CreateSeats(ctx context.Context, tx pgx.Tx, reservationID uuid.UUID, seatIDs []uuid.UUID) error {
	for _, seatID := range seatIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO reservation_seats (reservation_id, seat_id)
			VALUES ($1, $2)
		`, reservationID, seatID); err != nil {
			return sqlerr.HandleError(err)
		}
	}
	return nil
}

func (r *ReservationRepository) GetByID(ctx context.Context, reservationID uuid.UUID) (*model.Reservation, error) {
	var reservation model.Reservation
	if err := r.pool.QueryRow(ctx, `
		SELECT r.id, r.user_id, r.showtime_id, r.status, r.total_price, r.expires_at, r.created_at, r.updated_at,
		       m.title, sc.name, st.start_time, u.email, u.name
		FROM reservations r
		JOIN showtimes st ON st.id = r.showtime_id
		JOIN movies m ON m.id = st.movie_id
		JOIN screens sc ON sc.id = st.screen_id
		JOIN users u ON u.id = r.user_id
		WHERE r.id = $1
	`, reservationID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ShowtimeID,
		&reservation.Status,
		&reservation.TotalPrice,
		&reservation.ExpiresAt,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.MovieTitle,
		&reservation.ScreenName,
		&reservation.ShowStart,
		&reservation.UserEmail,
		&reservation.UserName,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	seats, err := r.GetSeats(ctx, reservation.ID)
	if err != nil {
		return nil, err
	}
	reservation.Seats = seats
	return &reservation, nil
}

func (r *ReservationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Reservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.user_id, r.showtime_id, r.status, r.total_price, r.expires_at, r.created_at, r.updated_at,
		       m.title, sc.name, st.start_time
		FROM reservations r
		JOIN showtimes st ON st.id = r.showtime_id
		JOIN movies m ON m.id = st.movie_id
		JOIN screens sc ON sc.id = st.screen_id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`, userID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var reservations []*model.Reservation
	for rows.Next() {
		var reservation model.Reservation
		if err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.ShowtimeID,
			&reservation.Status,
			&reservation.TotalPrice,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.MovieTitle,
			&reservation.ScreenName,
			&reservation.ShowStart,
		); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		seats, err := r.GetSeats(ctx, reservation.ID)
		if err != nil {
			return nil, err
		}
		reservation.Seats = seats
		reservations = append(reservations, &reservation)
	}
	return reservations, nil
}

func (r *ReservationRepository) Cancel(ctx context.Context, tx pgx.Tx, reservationID, userID uuid.UUID) (*model.Reservation, error) {
	var reservation model.Reservation
	if err := tx.QueryRow(ctx, `
		UPDATE reservations
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status IN ('pending', 'confirmed')
		RETURNING id, user_id, showtime_id, status, total_price, expires_at, created_at, updated_at
	`, reservationID, userID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ShowtimeID,
		&reservation.Status,
		&reservation.TotalPrice,
		&reservation.ExpiresAt,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &reservation, nil
}

func (r *ReservationRepository) GetSeatCountTx(ctx context.Context, tx pgx.Tx, reservationID uuid.UUID) (int, error) {
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM reservation_seats WHERE reservation_id = $1`, reservationID).Scan(&count); err != nil {
		return 0, sqlerr.HandleError(err)
	}
	return count, nil
}

func (r *ReservationRepository) GetSeats(ctx context.Context, reservationID uuid.UUID) ([]model.Seat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.screen_id, s.row, s.number, s.created_at, s.updated_at
		FROM seats s
		JOIN reservation_seats rs ON rs.seat_id = s.id
		WHERE rs.reservation_id = $1
		ORDER BY s.row ASC, s.number ASC
	`, reservationID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var seats []model.Seat
	for rows.Next() {
		var seat model.Seat
		if err := rows.Scan(&seat.ID, &seat.ScreenID, &seat.Row, &seat.Number, &seat.CreatedAt, &seat.UpdatedAt); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		seats = append(seats, seat)
	}
	return seats, nil
}

func (r *ReservationRepository) ListAll(ctx context.Context) ([]*model.Reservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.user_id, r.showtime_id, r.status, r.total_price, r.expires_at, r.created_at, r.updated_at,
		       m.title, sc.name, st.start_time, u.email, u.name
		FROM reservations r
		JOIN showtimes st ON st.id = r.showtime_id
		JOIN movies m ON m.id = st.movie_id
		JOIN screens sc ON sc.id = st.screen_id
		JOIN users u ON u.id = r.user_id
		ORDER BY r.created_at DESC
	`)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var reservations []*model.Reservation
	for rows.Next() {
		var reservation model.Reservation
		if err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.ShowtimeID,
			&reservation.Status,
			&reservation.TotalPrice,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.MovieTitle,
			&reservation.ScreenName,
			&reservation.ShowStart,
			&reservation.UserEmail,
			&reservation.UserName,
		); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		reservations = append(reservations, &reservation)
	}
	return reservations, nil
}

type RevenueRow struct {
	MovieID      uuid.UUID       `json:"movie_id"`
	MovieTitle   string          `json:"movie_title"`
	Count        int             `json:"count"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
}

func (r *ReservationRepository) RevenueSummary(ctx context.Context, from, to time.Time) ([]RevenueRow, decimal.Decimal, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.id, m.title, COUNT(r.id), COALESCE(SUM(r.total_price), 0)
		FROM reservations r
		JOIN showtimes st ON st.id = r.showtime_id
		JOIN movies m ON m.id = st.movie_id
		WHERE r.status IN ('pending', 'confirmed')
		  AND r.created_at >= $1
		  AND r.created_at < $2
		GROUP BY m.id, m.title
		ORDER BY m.title ASC
	`, from, to)
	if err != nil {
		return nil, decimal.Zero, sqlerr.HandleError(err)
	}
	defer rows.Close()

	rowsOut := make([]RevenueRow, 0)
	total := decimal.Zero
	for rows.Next() {
		var row RevenueRow
		if err := rows.Scan(&row.MovieID, &row.MovieTitle, &row.Count, &row.TotalRevenue); err != nil {
			return nil, decimal.Zero, sqlerr.HandleError(err)
		}
		total = total.Add(row.TotalRevenue)
		rowsOut = append(rowsOut, row)
	}
	return rowsOut, total, nil
}

type CapacityRow struct {
	ShowtimeID     uuid.UUID `json:"showtime_id"`
	MovieTitle     string    `json:"movie_title"`
	ScreenName     string    `json:"screen_name"`
	AvailableSeats int       `json:"available_seats"`
	ReservedSeats  int       `json:"reserved_seats"`
}

func (r *ReservationRepository) CapacitySummary(ctx context.Context) ([]CapacityRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT st.id, m.title, sc.name, st.available_seats, sc.total_seats - st.available_seats
		FROM showtimes st
		JOIN movies m ON m.id = st.movie_id
		JOIN screens sc ON sc.id = st.screen_id
		ORDER BY st.start_time ASC
	`)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	out := make([]CapacityRow, 0)
	for rows.Next() {
		var row CapacityRow
		if err := rows.Scan(&row.ShowtimeID, &row.MovieTitle, &row.ScreenName, &row.AvailableSeats, &row.ReservedSeats); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		out = append(out, row)
	}
	return out, nil
}

func (r *ReservationRepository) ExpirePending(ctx context.Context, tx pgx.Tx, before time.Time) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
		UPDATE reservations
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at < $1
		RETURNING id
	`, before)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *ReservationRepository) ReleasedSeatsByShowtime(ctx context.Context, tx pgx.Tx, reservationIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	rows, err := tx.Query(ctx, `
		SELECT r.showtime_id, COUNT(rs.seat_id)
		FROM reservations r
		JOIN reservation_seats rs ON rs.reservation_id = r.id
		WHERE r.id = ANY($1)
		GROUP BY r.showtime_id
	`, reservationIDs)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID]int)
	for rows.Next() {
		var showtimeID uuid.UUID
		var count int
		if err := rows.Scan(&showtimeID, &count); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		out[showtimeID] = count
	}
	return out, nil
}
