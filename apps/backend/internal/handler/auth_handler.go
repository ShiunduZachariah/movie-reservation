package handler

import (
	"net/http"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	base *Base
	auth *service.AuthService
}

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Username string `json:"username" validate:"required"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func NewAuthHandler(base *Base, auth *service.AuthService) *AuthHandler {
	return &AuthHandler{base: base, auth: auth}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := h.base.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.auth.Register(c.Request().Context(), req.Email, req.Password, req.Username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := h.base.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.auth.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, result)
}
