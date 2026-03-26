package handler

import (
	"net/http"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	base      *Base
	admin     *service.AdminService
	users     *service.UserService
	movies    *service.MovieService
	showtimes *service.ShowtimeService
}

func NewAdminHandler(base *Base, admin *service.AdminService, users *service.UserService, movies *service.MovieService, showtimes *service.ShowtimeService) *AdminHandler {
	return &AdminHandler{base: base, admin: admin, users: users, movies: movies, showtimes: showtimes}
}

func (h *AdminHandler) ListReservations(c echo.Context) error {
	reservations, err := h.admin.ListReservations(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"reservations": reservations})
}

func (h *AdminHandler) Revenue(c echo.Context) error {
	rows, total, err := h.admin.RevenueSummary(c.Request().Context(), c.QueryParam("from"), c.QueryParam("to"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"period": c.QueryParam("from") + " to " + c.QueryParam("to"),
		"rows":   rows,
		"totals": map[string]string{"revenue": total},
	})
}

func (h *AdminHandler) Capacity(c echo.Context) error {
	rows, err := h.admin.CapacitySummary(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"rows": rows})
}

func (h *AdminHandler) Promote(c echo.Context) error {
	if err := h.users.Promote(c.Request().Context(), c.Param("id")); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AdminHandler) Screens(c echo.Context) error {
	screens, err := h.showtimes.ListScreens(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"screens": screens})
}
