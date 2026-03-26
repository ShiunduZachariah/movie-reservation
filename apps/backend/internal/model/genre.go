package model

import "github.com/google/uuid"

type Genre struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
}
