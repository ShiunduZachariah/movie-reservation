package handler

import (
	"io"
	"net/http"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MovieHandler struct {
	base   *Base
	movies *service.MovieService
}

type saveMovieRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     string   `json:"description" validate:"required"`
	PosterURL       *string  `json:"poster_url"`
	DurationMinutes int      `json:"duration_minutes" validate:"required,min=1"`
	GenreIDs        []string `json:"genre_ids" validate:"required,min=1,dive,uuid"`
}

func NewMovieHandler(base *Base, movies *service.MovieService) *MovieHandler {
	return &MovieHandler{base: base, movies: movies}
}

func (h *MovieHandler) List(c echo.Context) error {
	movies, err := h.movies.List(c.Request().Context(), c.QueryParam("search"), c.QueryParam("genre_id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"movies": movies})
}

func (h *MovieHandler) Get(c echo.Context) error {
	movie, err := h.movies.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) Create(c echo.Context) error {
	var req saveMovieRequest
	if err := h.base.BindAndValidate(c, &req); err != nil {
		return err
	}
	input, err := parseSaveMovieInput(req)
	if err != nil {
		return err
	}
	movie, err := h.movies.Create(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, movie)
}

func (h *MovieHandler) Update(c echo.Context) error {
	var req saveMovieRequest
	if err := h.base.BindAndValidate(c, &req); err != nil {
		return err
	}
	input, err := parseSaveMovieInput(req)
	if err != nil {
		return err
	}
	movie, err := h.movies.Update(c.Request().Context(), c.Param("id"), input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) Delete(c echo.Context) error {
	if err := h.movies.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MovieHandler) UploadPoster(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	url, err := h.movies.UploadPoster(c.Request().Context(), c.Param("id"), data, file.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"poster_url": url})
}

func parseSaveMovieInput(req saveMovieRequest) (service.SaveMovieInput, error) {
	genreIDs := make([]uuid.UUID, 0, len(req.GenreIDs))
	for _, raw := range req.GenreIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return service.SaveMovieInput{}, errs.BadRequest("INVALID_GENRE_ID", "invalid genre id", nil)
		}
		genreIDs = append(genreIDs, id)
	}

	return service.SaveMovieInput{
		Title:           req.Title,
		Description:     req.Description,
		PosterURL:       req.PosterURL,
		DurationMinutes: req.DurationMinutes,
		GenreIDs:        genreIDs,
	}, nil
}
