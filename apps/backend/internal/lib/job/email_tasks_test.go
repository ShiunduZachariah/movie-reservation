package job

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTicketConfirmationTask(t *testing.T) {
	task, err := NewTicketConfirmationTask(TicketConfirmationPayload{
		UserEmail:      "zac@example.com",
		UserName:       "Zac",
		ReservationID:  "abc123",
		MovieTitle:     "The Matrix",
		ShowDate:       "Friday, 5 April 2026",
		ShowTime:       "7:30 PM",
		ScreenName:     "Screen 1",
		Seats:          []string{"A1", "A2"},
		TotalPrice:     "1200.00",
		ReservationURL: "http://localhost:3000/reservations/abc123",
	})
	require.NoError(t, err)
	require.Equal(t, TypeTicketConfirmation, task.Type())

	var payload TicketConfirmationPayload
	require.NoError(t, json.Unmarshal(task.Payload(), &payload))
	require.Equal(t, "zac@example.com", payload.UserEmail)
	require.Equal(t, []string{"A1", "A2"}, payload.Seats)
}
