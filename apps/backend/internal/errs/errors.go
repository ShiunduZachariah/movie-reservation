package errs

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type HTTPError struct {
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Status   int          `json:"status"`
	Override bool         `json:"override"`
	Errors   []FieldError `json:"errors,omitempty"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func New(status int, code, message string, override bool, fieldErrors []FieldError) *HTTPError {
	return &HTTPError{
		Code:     code,
		Message:  message,
		Status:   status,
		Override: override,
		Errors:   fieldErrors,
	}
}

func Unauthorized(code, message string) *HTTPError {
	return New(http.StatusUnauthorized, code, message, false, nil)
}

func Forbidden(code, message string) *HTTPError {
	return New(http.StatusForbidden, code, message, false, nil)
}

func NotFound(code, message string) *HTTPError {
	return New(http.StatusNotFound, code, message, false, nil)
}

func Conflict(code, message string) *HTTPError {
	return New(http.StatusConflict, code, message, false, nil)
}

func BadRequest(code, message string, fieldErrors []FieldError) *HTTPError {
	return New(http.StatusBadRequest, code, message, true, fieldErrors)
}

func Internal(message string) *HTTPError {
	return New(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, false, nil)
}

func Write(c echo.Context, err error) error {
	if httpErr, ok := err.(*HTTPError); ok {
		return c.JSON(httpErr.Status, httpErr)
	}
	return c.JSON(http.StatusInternalServerError, Internal("internal server error"))
}
