package handler

import (
	"net/http"
	"time"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/utils"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

type ShowtimeHandler struct {
	base      *Base
	showtimes *service.ShowtimeService
}

type createShowtimeRequest struct {
	MovieID     string `json:"movie_id" validate:"required,uuid"`
	ScreenID    string `json:"screen_id" validate:"required,uuid"`
	StartTime   string `json:"start_time" validate:"required"`
	EndTime     string `json:"end_time" validate:"required"`
	TicketPrice string `json:"ticket_price" validate:"required"`
}

func NewShowtimeHandler(base *Base, showtimes *service.ShowtimeService) *ShowtimeHandler {
	return &ShowtimeHandler{base: base, showtimes: showtimes}
}

func (h *ShowtimeHandler) Create(c echo.Context) error {
	var req createShowtimeRequest
	if err := h.base.BindAndValidate(c, &req); err != nil {
		return err
	}

	movieID, err := uuid.Parse(req.MovieID)
	if err != nil {
		return errs.BadRequest("INVALID_MOVIE_ID", "invalid movie id", nil)
	}
	screenID, err := uuid.Parse(req.ScreenID)
	if err != nil {
		return errs.BadRequest("INVALID_SCREEN_ID", "invalid screen id", nil)
	}
	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return errs.BadRequest("INVALID_START_TIME", "start_time must be RFC3339", nil)
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return errs.BadRequest("INVALID_END_TIME", "end_time must be RFC3339", nil)
	}
	price, err := decimal.NewFromString(req.TicketPrice)
	if err != nil {
		return errs.BadRequest("INVALID_TICKET_PRICE", "ticket_price must be decimal", nil)
	}

	showtime, err := h.showtimes.Create(c.Request().Context(), service.CreateShowtimeInput{
		MovieID:     movieID,
		ScreenID:    screenID,
		StartTime:   start,
		EndTime:     end,
		TicketPrice: price,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, showtime)
}

func (h *ShowtimeHandler) List(c echo.Context) error {
	date := c.QueryParam("date")
	movieID := c.QueryParam("movie_id")

	var (
		showtimes any
		err       error
	)
	switch {
	case date != "":
		showtimes, err = h.showtimes.ListByDate(c.Request().Context(), date)
	case movieID != "":
		showtimes, err = h.showtimes.ListByMovie(c.Request().Context(), movieID)
	default:
		return errs.BadRequest("DATE_OR_MOVIE_REQUIRED", "provide either date or movie_id", nil)
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"showtimes": showtimes})
}

func (h *ShowtimeHandler) Seats(c echo.Context) error {
	seats, err := h.showtimes.AvailableSeats(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	response := make([]map[string]any, 0, len(seats))
	for _, seat := range seats {
		response = append(response, map[string]any{
			"id":     seat.ID.String(),
			"row":    seat.Row,
			"number": seat.Number,
			"label":  utils.SeatLabel(seat.Row, seat.Number),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"showtime_id":     c.Param("id"),
		"available_count": len(response),
		"seats":           response,
	})
}

func (h *ShowtimeHandler) ScreenSeats(c echo.Context) error {
	seats, err := h.showtimes.ListSeatsByScreen(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	response := make([]map[string]any, 0, len(seats))
	for _, seat := range seats {
		response = append(response, map[string]any{
			"id":     seat.ID.String(),
			"row":    seat.Row,
			"number": seat.Number,
			"label":  utils.SeatLabel(seat.Row, seat.Number),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"screen_id": c.Param("id"),
		"count":     len(response),
		"seats":     response,
	})
}
