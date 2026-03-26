package model

import (
	"fmt"

	"github.com/google/uuid"
)

type Seat struct {
	Base
	ScreenID uuid.UUID `db:"screen_id" json:"screen_id"`
	Row      string    `db:"row" json:"row"`
	Number   int       `db:"number" json:"number"`
}

func (s Seat) Label() string {
	return s.Row + formatSeatNumber(s.Number)
}

func formatSeatNumber(number int) string {
	return fmt.Sprintf("%d", number)
}
