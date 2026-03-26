package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/resend/resend-go/v2"
)

type Client struct {
	resend      *resend.Client
	fromAddress string
}

type TicketConfirmationData struct {
	UserName       string
	ReservationID  string
	MovieTitle     string
	ShowDate       string
	ShowTime       string
	ScreenName     string
	Seats          []string
	TotalPrice     string
	ReservationURL string
}

func New(apiKey, fromAddress string) *Client {
	if fromAddress == "" {
		fromAddress = "CineReserve <onboarding@resend.dev>"
	}
	return &Client{
		resend:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
	}
}

func (c *Client) SendTicketConfirmation(ctx context.Context, to string, data TicketConfirmationData) error {
	html, err := c.renderTemplate("ticket_confirmation.html", data)
	if err != nil {
		return fmt.Errorf("render ticket confirmation template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    c.fromAddress,
		To:      []string{to},
		Subject: fmt.Sprintf("Your booking is confirmed - %s", data.MovieTitle),
		Html:    html,
	}

	if _, err := c.resend.Emails.SendWithContext(ctx, params); err != nil {
		return fmt.Errorf("send ticket confirmation: %w", err)
	}
	return nil
}

func (c *Client) renderTemplate(name string, data any) (string, error) {
	candidates := []string{
		filepath.Join("apps", "backend", "templates", "emails", name),
		filepath.Join("templates", "emails", name),
	}

	var lastErr error
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			lastErr = err
			continue
		}

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

	return "", fmt.Errorf("resolve template %s: %w", name, lastErr)
}
