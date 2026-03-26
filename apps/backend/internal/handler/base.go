package handler

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

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
		return translateValidationError(err, dest)
	}
	return nil
}

func translateValidationError(err error, dest any) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return errs.BadRequest("VALIDATION_ERROR", "request validation failed", nil)
	}

	fieldErrors := make([]errs.FieldError, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		fieldErrors = append(fieldErrors, errs.FieldError{
			Field: resolveJSONFieldName(dest, validationErr.StructField()),
			Error: humanizeValidationError(validationErr),
		})
	}

	return errs.BadRequest("VALIDATION_ERROR", fieldErrors[0].Error, fieldErrors)
}

func resolveJSONFieldName(dest any, structField string) string {
	t := reflect.TypeOf(dest)
	if t == nil {
		return strings.ToLower(structField)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return strings.ToLower(structField)
	}

	if field, ok := t.FieldByName(structField); ok {
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			name := strings.Split(jsonTag, ",")[0]
			if name != "" && name != "-" {
				return name
			}
		}
	}

	return strings.ToLower(structField)
}

func humanizeValidationError(err validator.FieldError) string {
	field := resolveValidationField(err)

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if err.Kind() == reflect.Slice || err.Kind() == reflect.Array {
			return fmt.Sprintf("%s must contain at least %s item(s)", field, err.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		if err.Kind() == reflect.Slice || err.Kind() == reflect.Array {
			return fmt.Sprintf("%s can contain at most %s item(s)", field, err.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, err.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func resolveValidationField(err validator.FieldError) string {
	if field := err.Field(); field != "" {
		return toSnakeCase(field)
	}
	return "field"
}

func toSnakeCase(value string) string {
	var builder strings.Builder
	for i, r := range value {
		if i > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteByte('_')
		}
		builder.WriteRune(r)
	}
	return strings.ToLower(builder.String())
}

func JSON(c echo.Context, status int, payload any) error {
	return c.JSON(status, payload)
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
