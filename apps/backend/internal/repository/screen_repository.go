package repository

import (
	"context"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/sqlerr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScreenRepository struct {
	pool *pgxpool.Pool
}

func NewScreenRepository(pool *pgxpool.Pool) *ScreenRepository {
	return &ScreenRepository{pool: pool}
}

func (r *ScreenRepository) List(ctx context.Context) ([]*model.Screen, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, total_seats, created_at, updated_at
		FROM screens
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var screens []*model.Screen
	for rows.Next() {
		var screen model.Screen
		if err := rows.Scan(&screen.ID, &screen.Name, &screen.TotalSeats, &screen.CreatedAt, &screen.UpdatedAt); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		screens = append(screens, &screen)
	}
	return screens, nil
}

func (r *ScreenRepository) GetByID(ctx context.Context, screenID uuid.UUID) (*model.Screen, error) {
	var screen model.Screen
	if err := r.pool.QueryRow(ctx, `
		SELECT id, name, total_seats, created_at, updated_at
		FROM screens
		WHERE id = $1
	`, screenID).Scan(&screen.ID, &screen.Name, &screen.TotalSeats, &screen.CreatedAt, &screen.UpdatedAt); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &screen, nil
}
