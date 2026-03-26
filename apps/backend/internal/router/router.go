package router

import (
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/handler"
	custommiddleware "github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/middleware"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/labstack/echo/v4"
)

func New(cfg *config.Config, handlers *handler.Handlers, services *service.Services) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = custommiddleware.HTTPErrorHandler
	e.Use(
		echoMiddleware.Recover(),
		echoMiddleware.RequestID(),
		echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
			AllowOrigins: cfg.Server.CORSAllowedOrigins,
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
		}),
	)

	auth := custommiddleware.NewAuthMiddleware(cfg.Auth.SecretKey, services.Users)

	e.GET("/status", handlers.Health.Status)

	api := e.Group("/api/v1")
	api.GET("/movies", handlers.Movies.List)
	api.GET("/movies/:id", handlers.Movies.Get)
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
	admin.GET("/admin/reservations", handlers.Admin.ListReservations)
	admin.GET("/admin/reports/revenue", handlers.Admin.Revenue)
	admin.GET("/admin/reports/capacity", handlers.Admin.Capacity)
	admin.PUT("/admin/users/:id/promote", handlers.Admin.Promote)

	return e
}
