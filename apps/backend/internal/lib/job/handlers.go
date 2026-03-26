package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/email"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

type Handlers struct {
	emailClient *email.Client
	logger      *zerolog.Logger
}

func NewHandlers(emailClient *email.Client, logger *zerolog.Logger) *Handlers {
	return &Handlers{emailClient: emailClient, logger: logger}
}

func (h *Handlers) HandleTicketConfirmation(ctx context.Context, task *asynq.Task) error {
	var payload TicketConfirmationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal ticket confirmation payload: %w", asynq.SkipRetry)
	}

	if err := h.emailClient.SendTicketConfirmation(ctx, payload.UserEmail, email.TicketConfirmationData{
		UserName:       payload.UserName,
		ReservationID:  payload.ReservationID,
		MovieTitle:     payload.MovieTitle,
		ShowDate:       payload.ShowDate,
		ShowTime:       payload.ShowTime,
		ScreenName:     payload.ScreenName,
		Seats:          payload.Seats,
		TotalPrice:     payload.TotalPrice,
		ReservationURL: payload.ReservationURL,
	}); err != nil {
		h.logger.Error().Err(err).Str("task", TypeTicketConfirmation).Msg("failed to send ticket confirmation email")
		return err
	}

	h.logger.Info().Str("task", TypeTicketConfirmation).Str("reservation_id", payload.ReservationID).Msg("ticket confirmation email sent")
	return nil
}
