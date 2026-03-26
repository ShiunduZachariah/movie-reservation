package middleware

import (
	"net/http"
	"strings"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	secretKey   string
	userService *service.UserService
}

func NewAuthMiddleware(secretKey string, userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{secretKey: secretKey, userService: userService}
}

func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return errs.Unauthorized("UNAUTHORIZED", "missing bearer token")
		}

		raw := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
			return []byte(m.secretKey), nil
		})
		if err != nil || !token.Valid {
			return errs.Unauthorized("UNAUTHORIZED", "invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return errs.Unauthorized("UNAUTHORIZED", "invalid token claims")
		}

		sub, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)
		name, _ := claims["name"].(string)
		if sub == "" || email == "" || name == "" {
			return errs.Unauthorized("UNAUTHORIZED", "missing token claims")
		}

		user, err := m.userService.SyncUser(c.Request().Context(), sub, email, name)
		if err != nil {
			return errs.Internal("failed to sync user")
		}
		c.Set("user_id", user.ID.String())
		c.Set("user_role", string(user.Role))
		return next(c)
	}
}

func (m *AuthMiddleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, _ := c.Get("user_role").(string)
		if role != "admin" {
			return errs.Forbidden("FORBIDDEN", "admin access required")
		}
		return next(c)
	}
}

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Response().Header().Get(echo.HeaderXRequestID) == "" {
				c.Response().Header().Set(echo.HeaderXRequestID, c.Request().Header.Get(echo.HeaderXRequestID))
			}
			return next(c)
		}
	}
}

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if httpErr, ok := err.(*errs.HTTPError); ok {
		_ = c.JSON(httpErr.Status, httpErr)
		return
	}

	echoErr, ok := err.(*echo.HTTPError)
	if ok {
		_ = c.JSON(echoErr.Code, map[string]any{"code": http.StatusText(echoErr.Code), "message": echoErr.Message, "status": echoErr.Code})
		return
	}

	_ = c.JSON(http.StatusInternalServerError, errs.Internal("internal server error"))
}
