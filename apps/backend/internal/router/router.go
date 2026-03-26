package router

import (
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/handler"
	custommiddleware "github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/middleware"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func New(cfg *config.Config, handlers *handler.Handlers, services *service.Services, logger *zerolog.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = custommiddleware.HTTPErrorHandler
	e.Use(
		echoMiddleware.Recover(),
		echoMiddleware.RequestID(),
		echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
			LogURI:      true,
			LogStatus:   true,
			LogMethod:   true,
			LogLatency:  true,
			LogRemoteIP: true,
			LogError:    true,
			LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
				entry := logger.Info().
					Str("event", "http_request").
					Str("method", v.Method).
					Str("uri", v.URI).
					Int("status", v.Status).
					Dur("latency", v.Latency).
					Str("remote_ip", v.RemoteIP)

				if requestID := c.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
					entry = entry.Str("request_id", requestID)
				}
				if userID, ok := c.Get("user_id").(string); ok && userID != "" {
					entry = entry.Str("user_id", userID)
				}
				if v.Error != nil {
					entry = entry.Err(v.Error)
				}
				entry.Msg("request completed")
				return nil
			},
		}),
		echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
			AllowOrigins: cfg.Server.CORSAllowedOrigins,
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
		}),
	)

	auth := custommiddleware.NewAuthMiddleware(cfg.Auth.SecretKey, services.Users)

	e.GET("/status", handlers.Health.Status)

	api := e.Group("/api/v1")
	api.POST("/auth/register", handlers.Auth.Register)
	api.POST("/auth/login", handlers.Auth.Login)
	api.GET("/genres", handlers.Movies.ListGenres)
	api.GET("/movies", handlers.Movies.List)
	api.GET("/movies/:id", handlers.Movies.Get)
	api.GET("/screens/:id/seats", handlers.Showtimes.ScreenSeats)
	api.GET("/showtimes", handlers.Showtimes.List)
	api.GET("/showtimes/:id/seats", handlers.Showtimes.Seats)

	authed := api.Group("", auth.RequireAuth)
	authed.POST("/reservations", handlers.Reservations.Create)
	authed.GET("/reservations/me", handlers.Reservations.ListMine)
	authed.GET("/reservations/:id", handlers.Reservations.Get)
	authed.DELETE("/reservations/:id", handlers.Reservations.Cancel)

	admin := authed.Group("", auth.RequireAdmin)
	admin.POST("/movies", handlers.Movies.Create)
	admin.PUT("/movies/:id", handlers.Movies.Update)
	admin.DELETE("/movies/:id", handlers.Movies.Delete)
	admin.POST("/movies/:id/poster", handlers.Movies.UploadPoster)
	admin.POST("/showtimes", handlers.Showtimes.Create)
	admin.GET("/admin/screens", handlers.Admin.Screens)
	admin.GET("/admin/reservations", handlers.Admin.ListReservations)
	admin.GET("/admin/reports/revenue", handlers.Admin.Revenue)
	admin.GET("/admin/reports/capacity", handlers.Admin.Capacity)
	admin.PUT("/admin/users/:id/promote", handlers.Admin.Promote)

	return e
}
