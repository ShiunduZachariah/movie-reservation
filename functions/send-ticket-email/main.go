package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/resend/resend-go/v2"
)

type TicketEmailMessage struct {
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
	ctx := context.Background()
	var msg TicketEmailMessage
	if err := json.NewDecoder(os.Stdin).Decode(&msg); err != nil {
		log.Fatalf("decode queue message: %v", err)
	}

	html, err := renderTemplate(msg)
	if err != nil {
		log.Fatalf("render template: %v", err)
	}

	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	params := &resend.SendEmailRequest{
		From:    "CineReserve <tickets@yourdomain.com>",
		To:      []string{msg.UserEmail},
		Subject: fmt.Sprintf("Your booking is confirmed - %s", msg.MovieTitle),
		Html:    html,
	}
	if _, err := client.Emails.SendWithContext(ctx, params); err != nil {
		log.Fatalf("send email: %v", err)
	}
	log.Printf("sent ticket confirmation to %s for reservation %s", msg.UserEmail, msg.ReservationID)
}

func renderTemplate(data TicketEmailMessage) (string, error) {
	tmpl, err := template.ParseFiles(filepath.Join("apps", "backend", "templates", "emails", "ticket_confirmation.html"))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
