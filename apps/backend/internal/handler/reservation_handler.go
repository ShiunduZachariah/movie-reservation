package handler

import (
	"net/http"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ReservationHandler struct {
	base         *Base
	reservations *service.ReservationService
}

type createReservationRequest struct {
	ShowtimeID string   `json:"showtime_id" validate:"required,uuid"`
	SeatIDs    []string `json:"seat_ids" validate:"required,min=1,max=10,dive,uuid"`
}

func NewReservationHandler(base *Base, reservations *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{base: base, reservations: reservations}
}

func (h *ReservationHandler) Create(c echo.Context) error {
	var req createReservationRequest
	if err := h.base.BindAndValidate(c, &req); err != nil {
		return err
	}
	showtimeID, err := uuid.Parse(req.ShowtimeID)
	if err != nil {
		return errs.BadRequest("INVALID_SHOWTIME_ID", "invalid showtime id", nil)
	}
	seatIDs := make([]uuid.UUID, 0, len(req.SeatIDs))
	for _, raw := range req.SeatIDs {
		seatID, err := uuid.Parse(raw)
		if err != nil {
			return errs.BadRequest("INVALID_SEAT_ID", "invalid seat id", nil)
		}
		seatIDs = append(seatIDs, seatID)
	}
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return errs.Unauthorized("UNAUTHORIZED", "missing authenticated user")
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return errs.Unauthorized("UNAUTHORIZED", "invalid authenticated user")
	}

	reservation, err := h.reservations.ReserveSeats(c.Request().Context(), service.CreateReservationInput{
		UserID:     parsedUserID,
		ShowtimeID: showtimeID,
		SeatIDs:    seatIDs,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, reservation)
}

func (h *ReservationHandler) ListMine(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	reservations, err := h.reservations.ListMyReservations(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"reservations": reservations})
}

func (h *ReservationHandler) Get(c echo.Context) error {
	reservation, err := h.reservations.GetReservation(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, reservation)
}

func (h *ReservationHandler) Cancel(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	reservation, err := h.reservations.CancelReservation(c.Request().Context(), c.Param("id"), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, reservation)
}
