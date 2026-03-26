package model

type Movie struct {
	Base
	Title           string  `db:"title" json:"title"`
	Description     string  `db:"description" json:"description"`
	PosterURL       *string `db:"poster_url" json:"poster_url,omitempty"`
	DurationMinutes int     `db:"duration_minutes" json:"duration_minutes"`
	Genres          []Genre `db:"-" json:"genres,omitempty"`
}
