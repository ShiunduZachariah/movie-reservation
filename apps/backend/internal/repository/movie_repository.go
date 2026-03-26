package repository

import (
	"context"
	"strconv"
	"strings"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/sqlerr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieRepository struct {
	pool *pgxpool.Pool
}

type MovieFilters struct {
	Search  string
	GenreID *uuid.UUID
}

func NewMovieRepository(pool *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{pool: pool}
}

func (r *MovieRepository) Create(ctx context.Context, db DBTX, movie *model.Movie) (*model.Movie, error) {
	var created model.Movie
	if err := db.QueryRow(ctx, `
		INSERT INTO movies (title, description, poster_url, duration_minutes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, poster_url, duration_minutes, created_at, updated_at
	`, movie.Title, movie.Description, movie.PosterURL, movie.DurationMinutes).Scan(
		&created.ID,
		&created.Title,
		&created.Description,
		&created.PosterURL,
		&created.DurationMinutes,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &created, nil
}

func (r *MovieRepository) Update(ctx context.Context, db DBTX, movieID uuid.UUID, movie *model.Movie) (*model.Movie, error) {
	var updated model.Movie
	if err := db.QueryRow(ctx, `
		UPDATE movies
		SET title = $2, description = $3, poster_url = $4, duration_minutes = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, description, poster_url, duration_minutes, created_at, updated_at
	`, movieID, movie.Title, movie.Description, movie.PosterURL, movie.DurationMinutes).Scan(
		&updated.ID,
		&updated.Title,
		&updated.Description,
		&updated.PosterURL,
		&updated.DurationMinutes,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &updated, nil
}

func (r *MovieRepository) Delete(ctx context.Context, db DBTX, movieID uuid.UUID) error {
	tag, err := db.Exec(ctx, `DELETE FROM movies WHERE id = $1`, movieID)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	if tag.RowsAffected() == 0 {
		return sqlerr.HandleError(pgx.ErrNoRows)
	}
	return nil
}

func (r *MovieRepository) GetByID(ctx context.Context, movieID uuid.UUID) (*model.Movie, error) {
	var movie model.Movie
	if err := r.pool.QueryRow(ctx, `
		SELECT id, title, description, poster_url, duration_minutes, created_at, updated_at
		FROM movies
		WHERE id = $1
	`, movieID).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.PosterURL,
		&movie.DurationMinutes,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	genres, err := r.GetGenres(ctx, movieID)
	if err != nil {
		return nil, err
	}
	movie.Genres = genres
	return &movie, nil
}

func (r *MovieRepository) List(ctx context.Context, filters MovieFilters) ([]*model.Movie, error) {
	var args []any
	var conditions []string
	query := strings.Builder{}
	query.WriteString(`
		SELECT DISTINCT m.id, m.title, m.description, m.poster_url, m.duration_minutes, m.created_at, m.updated_at
		FROM movies m
		LEFT JOIN movie_genres mg ON mg.movie_id = m.id
	`)

	if filters.Search != "" {
		args = append(args, "%"+filters.Search+"%")
		conditions = append(conditions, "m.title ILIKE $"+strconv.Itoa(len(args)))
	}
	if filters.GenreID != nil {
		args = append(args, *filters.GenreID)
		conditions = append(conditions, "mg.genre_id = $"+strconv.Itoa(len(args)))
	}
	if len(conditions) > 0 {
		query.WriteString(" WHERE " + strings.Join(conditions, " AND "))
	}
	query.WriteString(" ORDER BY m.created_at DESC")

	rows, err := r.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var movies []*model.Movie
	for rows.Next() {
		var movie model.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.PosterURL, &movie.DurationMinutes, &movie.CreatedAt, &movie.UpdatedAt); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		genres, err := r.GetGenres(ctx, movie.ID)
		if err != nil {
			return nil, err
		}
		movie.Genres = genres
		movies = append(movies, &movie)
	}
	return movies, nil
}

func (r *MovieRepository) AttachGenres(ctx context.Context, db DBTX, movieID uuid.UUID, genreIDs []uuid.UUID) error {
	if _, err := db.Exec(ctx, `DELETE FROM movie_genres WHERE movie_id = $1`, movieID); err != nil {
		return sqlerr.HandleError(err)
	}

	for _, genreID := range genreIDs {
		if _, err := db.Exec(ctx, `
			INSERT INTO movie_genres (movie_id, genre_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, movieID, genreID); err != nil {
			return sqlerr.HandleError(err)
		}
	}
	return nil
}

func (r *MovieRepository) GetGenres(ctx context.Context, movieID uuid.UUID) ([]model.Genre, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.name
		FROM genres g
		JOIN movie_genres mg ON mg.genre_id = g.id
		WHERE mg.movie_id = $1
		ORDER BY g.name ASC
	`, movieID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var genres []model.Genre
	for rows.Next() {
		var genre model.Genre
		if err := rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		genres = append(genres, genre)
	}
	return genres, nil
}

func (r *MovieRepository) UpdatePosterURL(ctx context.Context, db DBTX, movieID uuid.UUID, posterURL string) error {
	tag, err := db.Exec(ctx, `UPDATE movies SET poster_url = $2, updated_at = NOW() WHERE id = $1`, movieID, posterURL)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	if tag.RowsAffected() == 0 {
		return sqlerr.HandleError(pgx.ErrNoRows)
	}
	return nil
}
