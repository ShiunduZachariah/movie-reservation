package handler

import "github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"

type Handlers struct {
	Base         *Base
	Health       *HealthHandler
	Auth         *AuthHandler
	Movies       *MovieHandler
	Showtimes    *ShowtimeHandler
	Reservations *ReservationHandler
	Admin        *AdminHandler
}

func New(services *service.Services) *Handlers {
	base := NewBase()
	return &Handlers{
		Base:         base,
		Health:       NewHealthHandler(),
		Auth:         NewAuthHandler(base, services.Auth),
		Movies:       NewMovieHandler(base, services.Movies),
		Showtimes:    NewShowtimeHandler(base, services.Showtimes),
		Reservations: NewReservationHandler(base, services.Reservations),
		Admin:        NewAdminHandler(base, services.Admin, services.Users, services.Movies, services.Showtimes),
	}
}
