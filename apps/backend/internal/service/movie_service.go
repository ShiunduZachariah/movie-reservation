package service

import (
	"context"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/blob"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieService struct {
	db            *pgxpool.Pool
	movies        *repository.MovieRepository
	genres        *repository.GenreRepository
	blobClient    *blob.Client
	containerName string
}

type SaveMovieInput struct {
	Title           string
	Description     string
	PosterURL       *string
	DurationMinutes int
	GenreIDs        []uuid.UUID
}

func NewMovieService(db *pgxpool.Pool, movies *repository.MovieRepository, genres *repository.GenreRepository, blobClient *blob.Client, containerName string) *MovieService {
	return &MovieService{db: db, movies: movies, genres: genres, blobClient: blobClient, containerName: containerName}
}

func (s *MovieService) List(ctx context.Context, search, genreID string) ([]*model.Movie, error) {
	filters := repository.MovieFilters{Search: search}
	if genreID != "" {
		id, err := uuid.Parse(genreID)
		if err != nil {
			return nil, errs.BadRequest("INVALID_GENRE_ID", "invalid genre id", nil)
		}
		filters.GenreID = &id
	}
	return s.movies.List(ctx, filters)
}

func (s *MovieService) Get(ctx context.Context, movieID string) (*model.Movie, error) {
	id, err := uuid.Parse(movieID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_MOVIE_ID", "invalid movie id", nil)
	}
	return s.movies.GetByID(ctx, id)
}

func (s *MovieService) Create(ctx context.Context, input SaveMovieInput) (*model.Movie, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	movie, err := s.movies.Create(ctx, tx, &model.Movie{
		Title:           input.Title,
		Description:     input.Description,
		PosterURL:       input.PosterURL,
		DurationMinutes: input.DurationMinutes,
	})
	if err != nil {
		return nil, err
	}
	if err := s.movies.AttachGenres(ctx, tx, movie.ID, input.GenreIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.movies.GetByID(ctx, movie.ID)
}

func (s *MovieService) Update(ctx context.Context, movieID string, input SaveMovieInput) (*model.Movie, error) {
	id, err := uuid.Parse(movieID)
	if err != nil {
		return nil, errs.BadRequest("INVALID_MOVIE_ID", "invalid movie id", nil)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := s.movies.Update(ctx, tx, id, &model.Movie{
		Title:           input.Title,
		Description:     input.Description,
		PosterURL:       input.PosterURL,
		DurationMinutes: input.DurationMinutes,
	}); err != nil {
		return nil, err
	}
	if err := s.movies.AttachGenres(ctx, tx, id, input.GenreIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.movies.GetByID(ctx, id)
}

func (s *MovieService) Delete(ctx context.Context, movieID string) error {
	id, err := uuid.Parse(movieID)
	if err != nil {
		return errs.BadRequest("INVALID_MOVIE_ID", "invalid movie id", nil)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.movies.Delete(ctx, tx, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *MovieService) UploadPoster(ctx context.Context, movieID string, data []byte, contentType string) (string, error) {
	if s.blobClient == nil {
		return "", errs.Internal("blob storage not configured")
	}
	id, err := uuid.Parse(movieID)
	if err != nil {
		return "", errs.BadRequest("INVALID_MOVIE_ID", "invalid movie id", nil)
	}

	url, err := s.blobClient.UploadPoster(ctx, s.containerName, id.String(), data, contentType)
	if err != nil {
		return "", err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	if err := s.movies.UpdatePosterURL(ctx, tx, id, url); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return url, nil
}
