package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("missing DATABASE_URL")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	expired, err := expirePending(ctx, pool, time.Now().UTC())
	if err != nil {
		log.Fatalf("expire pending reservations: %v", err)
	}
	log.Printf("expired %d reservations", len(expired))
}

func expirePending(ctx context.Context, pool *pgxpool.Pool, before time.Time) ([]uuid.UUID, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		UPDATE reservations
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at < $1
		RETURNING id
	`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if len(ids) > 0 {
		releaseRows, err := tx.Query(ctx, `
			SELECT r.showtime_id, COUNT(rs.seat_id)
			FROM reservations r
			JOIN reservation_seats rs ON rs.reservation_id = r.id
			WHERE r.id = ANY($1)
			GROUP BY r.showtime_id
		`, ids)
		if err != nil {
			return nil, err
		}
		defer releaseRows.Close()

		for releaseRows.Next() {
			var showtimeID uuid.UUID
			var count int
			if err := releaseRows.Scan(&showtimeID, &count); err != nil {
				return nil, err
			}
			if _, err := tx.Exec(ctx, `
				UPDATE showtimes
				SET available_seats = available_seats + $2, updated_at = NOW()
				WHERE id = $1
			`, showtimeID, count); err != nil {
				return nil, fmt.Errorf("restore seats for %s: %w", showtimeID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ids, nil
}
