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

type ShowtimeRepository struct {
	pool *pgxpool.Pool
}

func NewShowtimeRepository(pool *pgxpool.Pool) *ShowtimeRepository {
	return &ShowtimeRepository{pool: pool}
}

func (r *ShowtimeRepository) Create(ctx context.Context, db DBTX, showtime *model.Showtime) (*model.Showtime, error) {
	var created model.Showtime
	if err := db.QueryRow(ctx, `
		INSERT INTO showtimes (movie_id, screen_id, start_time, end_time, ticket_price, available_seats)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, movie_id, screen_id, start_time, end_time, ticket_price, available_seats, created_at, updated_at
	`, showtime.MovieID, showtime.ScreenID, showtime.StartTime, showtime.EndTime, showtime.TicketPrice, showtime.AvailableSeats).Scan(
		&created.ID,
		&created.MovieID,
		&created.ScreenID,
		&created.StartTime,
		&created.EndTime,
		&created.TicketPrice,
		&created.AvailableSeats,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &created, nil
}

func (r *ShowtimeRepository) GetByID(ctx context.Context, showtimeID uuid.UUID) (*model.Showtime, error) {
	var showtime model.Showtime
	if err := r.pool.QueryRow(ctx, `
		SELECT st.id, st.movie_id, st.screen_id, st.start_time, st.end_time, st.ticket_price, st.available_seats, st.created_at, st.updated_at,
		       m.title, sc.name
		FROM showtimes st
		JOIN movies m ON m.id = st.movie_id
		JOIN screens sc ON sc.id = st.screen_id
		WHERE st.id = $1
	`, showtimeID).Scan(
		&showtime.ID,
		&showtime.MovieID,
		&showtime.ScreenID,
		&showtime.StartTime,
		&showtime.EndTime,
		&showtime.TicketPrice,
		&showtime.AvailableSeats,
		&showtime.CreatedAt,
		&showtime.UpdatedAt,
		&showtime.MovieTitle,
		&showtime.ScreenName,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &showtime, nil
}

func (r *ShowtimeRepository) ListByDate(ctx context.Context, date time.Time) ([]*model.Showtime, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT st.id, st.movie_id, st.screen_id, st.start_time, st.end_time, st.ticket_price, st.available_seats, st.created_at, st.updated_at,
		       m.title, sc.name
		FROM showtimes st
		JOIN movies m ON m.id = st.movie_id
		JOIN screens sc ON sc.id = st.screen_id
		WHERE DATE(st.start_time) = DATE($1)
		ORDER BY st.start_time ASC
	`, date)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var showtimes []*model.Showtime
	for rows.Next() {
		var showtime model.Showtime
		if err := rows.Scan(
			&showtime.ID,
			&showtime.MovieID,
			&showtime.ScreenID,
			&showtime.StartTime,
			&showtime.EndTime,
			&showtime.TicketPrice,
			&showtime.AvailableSeats,
			&showtime.CreatedAt,
			&showtime.UpdatedAt,
			&showtime.MovieTitle,
			&showtime.ScreenName,
		); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		showtimes = append(showtimes, &showtime)
	}
	return showtimes, nil
}

func (r *ShowtimeRepository) ListByMovie(ctx context.Context, movieID uuid.UUID) ([]*model.Showtime, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT st.id, st.movie_id, st.screen_id, st.start_time, st.end_time, st.ticket_price, st.available_seats, st.created_at, st.updated_at,
		       m.title, sc.name
		FROM showtimes st
		JOIN movies m ON m.id = st.movie_id
		JOIN screens sc ON sc.id = st.screen_id
		WHERE st.movie_id = $1
		ORDER BY st.start_time ASC
	`, movieID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var showtimes []*model.Showtime
	for rows.Next() {
		var showtime model.Showtime
		if err := rows.Scan(
			&showtime.ID,
			&showtime.MovieID,
			&showtime.ScreenID,
			&showtime.StartTime,
			&showtime.EndTime,
			&showtime.TicketPrice,
			&showtime.AvailableSeats,
			&showtime.CreatedAt,
			&showtime.UpdatedAt,
			&showtime.MovieTitle,
			&showtime.ScreenName,
		); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		showtimes = append(showtimes, &showtime)
	}
	return showtimes, nil
}

func (r *ShowtimeRepository) GetForUpdate(ctx context.Context, tx pgx.Tx, showtimeID uuid.UUID) (*model.Showtime, error) {
	var showtime model.Showtime
	if err := tx.QueryRow(ctx, `
		SELECT id, movie_id, screen_id, start_time, end_time, ticket_price, available_seats, created_at, updated_at
		FROM showtimes
		WHERE id = $1
		FOR UPDATE
	`, showtimeID).Scan(
		&showtime.ID,
		&showtime.MovieID,
		&showtime.ScreenID,
		&showtime.StartTime,
		&showtime.EndTime,
		&showtime.TicketPrice,
		&showtime.AvailableSeats,
		&showtime.CreatedAt,
		&showtime.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &showtime, nil
}

func (r *ShowtimeRepository) DecrementAvailableSeats(ctx context.Context, tx pgx.Tx, showtimeID uuid.UUID, count int) error {
	tag, err := tx.Exec(ctx, `
		UPDATE showtimes
		SET available_seats = available_seats - $2, updated_at = NOW()
		WHERE id = $1 AND available_seats >= $2
	`, showtimeID, count)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	if tag.RowsAffected() == 0 {
		return sqlerr.HandleError(pgx.ErrNoRows)
	}
	return nil
}

func (r *ShowtimeRepository) IncrementAvailableSeats(ctx context.Context, tx DBTX, showtimeID uuid.UUID, count int) error {
	tag, err := tx.Exec(ctx, `
		UPDATE showtimes
		SET available_seats = available_seats + $2, updated_at = NOW()
		WHERE id = $1
	`, showtimeID, count)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	if tag.RowsAffected() == 0 {
		return sqlerr.HandleError(pgx.ErrNoRows)
	}
	return nil
}

func (r *ShowtimeRepository) GetTicketPrice(ctx context.Context, showtimeID uuid.UUID) (decimal.Decimal, error) {
	var price decimal.Decimal
	if err := r.pool.QueryRow(ctx, `SELECT ticket_price FROM showtimes WHERE id = $1`, showtimeID).Scan(&price); err != nil {
		return decimal.Zero, sqlerr.HandleError(err)
	}
	return price, nil
}
