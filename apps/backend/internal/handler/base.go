package handler

import (
	"net/http"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Base struct {
	validate *validator.Validate
}

func NewBase() *Base {
	return &Base{validate: validator.New()}
}

func (b *Base) BindAndValidate(c echo.Context, dest any) error {
	if err := c.Bind(dest); err != nil {
		return errs.BadRequest("INVALID_REQUEST", "invalid request payload", nil)
	}
	if err := b.validate.Struct(dest); err != nil {
		return errs.BadRequest("VALIDATION_ERROR", err.Error(), nil)
	}
	return nil
}

func JSON(c echo.Context, status int, payload any) error {
	return c.JSON(status, payload)
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
