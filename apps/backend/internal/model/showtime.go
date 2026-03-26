package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Showtime struct {
	Base
	MovieID        uuid.UUID       `db:"movie_id" json:"movie_id"`
	ScreenID       uuid.UUID       `db:"screen_id" json:"screen_id"`
	StartTime      time.Time       `db:"start_time" json:"start_time"`
	EndTime        time.Time       `db:"end_time" json:"end_time"`
	TicketPrice    decimal.Decimal `db:"ticket_price" json:"ticket_price"`
	AvailableSeats int             `db:"available_seats" json:"available_seats"`
	MovieTitle     string          `db:"movie_title" json:"movie_title,omitempty"`
	ScreenName     string          `db:"screen_name" json:"screen_name,omitempty"`
}
