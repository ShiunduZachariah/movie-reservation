package repository

import (
	"context"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/sqlerr"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GenreRepository struct {
	pool *pgxpool.Pool
}

func NewGenreRepository(pool *pgxpool.Pool) *GenreRepository {
	return &GenreRepository{pool: pool}
}

func (r *GenreRepository) ListAll(ctx context.Context) ([]model.Genre, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM genres ORDER BY name ASC`)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	genres := make([]model.Genre, 0)
	for rows.Next() {
		var genre model.Genre
		if err := rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		genres = append(genres, genre)
	}

	return genres, nil
}
