package model

type Screen struct {
	Base
	Name       string `db:"name" json:"name"`
	TotalSeats int    `db:"total_seats" json:"total_seats"`
}
