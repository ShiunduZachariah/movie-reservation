package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/resend/resend-go/v2"
)

const fromAddress = "CineReserve <tickets@yourdomain.com>"

type Client struct {
	resend *resend.Client
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

func New(apiKey string) *Client {
	return &Client{resend: resend.NewClient(apiKey)}
}

func (c *Client) SendTicketConfirmation(ctx context.Context, to string, data TicketConfirmationData) error {
	html, err := c.renderTemplate("ticket_confirmation.html", data)
	if err != nil {
		return fmt.Errorf("render ticket confirmation template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    fromAddress,
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
	tmpl, err := template.ParseFiles(filepath.Join("templates", "emails", name))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
