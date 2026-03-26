package service

import (
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/blob"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/email"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/job"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Services struct {
	Users        *UserService
	Movies       *MovieService
	Showtimes    *ShowtimeService
	Reservations *ReservationService
	Admin        *AdminService
}

func New(cfg *config.Config, repos *repository.Repositories, db *pgxpool.Pool, jobs *job.Service, emailClient *email.Client, blobClient *blob.Client, queueClient *azqueue.QueueClient, logger *zerolog.Logger) *Services {
	users := NewUserService(db, repos.Users)
	movies := NewMovieService(db, repos.Movies, repos.Genres, blobClient, cfg.Azure.StorageContainerName)
	showtimes := NewShowtimeService(db, repos.Showtimes, repos.Screens, repos.Seats)
	reservations := NewReservationService(cfg, db, repos, jobs, queueClient, logger)
	admin := NewAdminService(db, repos, users)

	return &Services{
		Users:        users,
		Movies:       movies,
		Showtimes:    showtimes,
		Reservations: reservations,
		Admin:        admin,
	}
}
