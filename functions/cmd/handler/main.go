package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resend/resend-go/v2"
)

type invocationRequest struct {
	Data map[string]json.RawMessage `json:"Data"`
}

type invocationResponse struct {
	Outputs     map[string]any `json:"Outputs,omitempty"`
	Logs        []string       `json:"Logs,omitempty"`
	ReturnValue any            `json:"ReturnValue,omitempty"`
}

type ticketEmailMessage struct {
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

func main() {
	defaultPort := os.Getenv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if defaultPort == "" {
		defaultPort = "8080"
	}

	port := flag.String("port", defaultPort, "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/send-ticket-email", handleSendTicketEmail)
	mux.HandleFunc("/expire-reservations", handleExpireReservations)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	})

	addr := ":" + *port
	log.Printf("starting custom handler on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("custom handler stopped: %v", err)
	}
}

func handleSendTicketEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request invocationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("decode invocation: %v", err), http.StatusBadRequest)
		return
	}

	raw, ok := request.Data["ticketEmailQueue"]
	if !ok {
		http.Error(w, "missing queue payload", http.StatusBadRequest)
		return
	}

	message, err := decodeTicketEmail(raw)
	if err != nil {
		http.Error(w, fmt.Sprintf("decode ticket email payload: %v", err), http.StatusBadRequest)
		return
	}

	if err := sendTicketEmail(r.Context(), message); err != nil {
		http.Error(w, fmt.Sprintf("send ticket email: %v", err), http.StatusInternalServerError)
		return
	}

	writeInvocationResponse(w, http.StatusOK, invocationResponse{
		Logs: []string{fmt.Sprintf("sent ticket confirmation to %s for reservation %s", message.UserEmail, message.ReservationID)},
	})
}

func handleExpireReservations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		http.Error(w, "missing DATABASE_URL", http.StatusInternalServerError)
		return
	}

	pool, err := pgxpool.New(r.Context(), dsn)
	if err != nil {
		http.Error(w, fmt.Sprintf("connect database: %v", err), http.StatusInternalServerError)
		return
	}
	defer pool.Close()

	expired, err := expirePending(r.Context(), pool, time.Now().UTC())
	if err != nil {
		http.Error(w, fmt.Sprintf("expire pending reservations: %v", err), http.StatusInternalServerError)
		return
	}

	writeInvocationResponse(w, http.StatusOK, invocationResponse{
		Logs: []string{fmt.Sprintf("expired %d reservations", len(expired))},
	})
}

func decodeTicketEmail(raw json.RawMessage) (ticketEmailMessage, error) {
	var message ticketEmailMessage

	var stringPayload string
	if err := json.Unmarshal(raw, &stringPayload); err == nil {
		if err := json.Unmarshal([]byte(stringPayload), &message); err != nil {
			return ticketEmailMessage{}, err
		}
		return message, nil
	}

	if err := json.Unmarshal(raw, &message); err != nil {
		return ticketEmailMessage{}, err
	}
	return message, nil
}

func sendTicketEmail(ctx context.Context, data ticketEmailMessage) error {
	html, err := renderTemplate(data)
	if err != nil {
		return err
	}

	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey == "" {
		return fmt.Errorf("missing RESEND_API_KEY")
	}

	fromAddress := os.Getenv("RESEND_FROM")
	if fromAddress == "" {
		fromAddress = "CineReserve <onboarding@resend.dev>"
	}

	client := resend.NewClient(resendAPIKey)
	params := &resend.SendEmailRequest{
		From:    fromAddress,
		To:      []string{data.UserEmail},
		Subject: fmt.Sprintf("Your booking is confirmed - %s", data.MovieTitle),
		Html:    html,
	}

	_, err = client.Emails.SendWithContext(ctx, params)
	return err
}

func renderTemplate(data ticketEmailMessage) (string, error) {
	candidates := []string{
		filepath.Join("templates", "emails", "ticket_confirmation.html"),
		filepath.Join("apps", "backend", "templates", "emails", "ticket_confirmation.html"),
	}

	var lastErr error
	for _, candidate := range candidates {
		tmpl, err := template.ParseFiles(candidate)
		if err != nil {
			lastErr = err
			continue
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	return "", fmt.Errorf("load email template: %w", lastErr)
}

func expirePending(ctx context.Context, pool *pgxpool.Pool, before time.Time) ([]uuid.UUID, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		UPDATE reservations
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at < $1
		RETURNING id
	`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if len(ids) > 0 {
		releaseRows, err := tx.Query(ctx, `
			SELECT r.showtime_id, COUNT(rs.seat_id)
			FROM reservations r
			JOIN reservation_seats rs ON rs.reservation_id = r.id
			WHERE r.id = ANY($1)
			GROUP BY r.showtime_id
		`, ids)
		if err != nil {
			return nil, err
		}
		defer releaseRows.Close()

		for releaseRows.Next() {
			var showtimeID uuid.UUID
			var count int
			if err := releaseRows.Scan(&showtimeID, &count); err != nil {
				return nil, err
			}
			if _, err := tx.Exec(ctx, `
				UPDATE showtimes
				SET available_seats = available_seats + $2, updated_at = NOW()
				WHERE id = $1
			`, showtimeID, count); err != nil {
				return nil, fmt.Errorf("restore seats for %s: %w", showtimeID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ids, nil
}

func writeInvocationResponse(w http.ResponseWriter, status int, payload invocationResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode invocation response: %v", err)
	}
}
