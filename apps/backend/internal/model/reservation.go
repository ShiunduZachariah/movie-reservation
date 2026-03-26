package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "pending"
	ReservationConfirmed ReservationStatus = "confirmed"
	ReservationCancelled ReservationStatus = "cancelled"
	ReservationExpired   ReservationStatus = "expired"
)

type Reservation struct {
	Base
	UserID     uuid.UUID         `db:"user_id" json:"user_id"`
	ShowtimeID uuid.UUID         `db:"showtime_id" json:"showtime_id"`
	Status     ReservationStatus `db:"status" json:"status"`
	TotalPrice decimal.Decimal   `db:"total_price" json:"total_price"`
	ExpiresAt  time.Time         `db:"expires_at" json:"expires_at"`
	Seats      []Seat            `db:"-" json:"seats,omitempty"`
	MovieTitle string            `db:"movie_title" json:"movie_title,omitempty"`
	ScreenName string            `db:"screen_name" json:"screen_name,omitempty"`
	ShowStart  time.Time         `db:"show_start" json:"show_start,omitempty"`
	UserEmail  string            `db:"user_email" json:"user_email,omitempty"`
	UserName   string            `db:"user_name" json:"user_name,omitempty"`
}
