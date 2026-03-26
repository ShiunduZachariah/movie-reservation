package repository

import "github.com/jackc/pgx/v5/pgxpool"

type Repositories struct {
	Users        *UserRepository
	Genres       *GenreRepository
	Movies       *MovieRepository
	Screens      *ScreenRepository
	Seats        *SeatRepository
	Showtimes    *ShowtimeRepository
	Reservations *ReservationRepository
}

func New(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Users:        NewUserRepository(pool),
		Genres:       NewGenreRepository(pool),
		Movies:       NewMovieRepository(pool),
		Screens:      NewScreenRepository(pool),
		Seats:        NewSeatRepository(pool),
		Showtimes:    NewShowtimeRepository(pool),
		Reservations: NewReservationRepository(pool),
	}
}
