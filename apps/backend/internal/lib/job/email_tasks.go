package job

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const TypeTicketConfirmation = "reservation:ticket_confirmation"

type TicketConfirmationPayload struct {
	UserEmail      string   `json:"user_email"`
	UserName       string   `json:"user_name"`
	ReservationID  string   `json:"reservation_id"`
	MovieTitle     string   `json:"movie_title"`
	ShowDate       string   `json:"show_date"`
	ShowTime       string   `json:"show_time"`
	ScreenName     string   `json:"screen_name"`
	Seats          []string `json:"seats"`
	TotalPrice     string   `json:"total_price"`
	ReservationURL string   `json:"reservation_url"`
}

func NewTicketConfirmationTask(p TicketConfirmationPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal ticket confirmation payload: %w", err)
	}
	return asynq.NewTask(
		TypeTicketConfirmation,
		payload,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
		asynq.Timeout(30*time.Second),
	), nil
}
