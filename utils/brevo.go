package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type brevoEmailRequest struct {
	Sender struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"sender"`
	To []struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	} `json:"to"`
	Subject     string `json:"subject"`
	HTMLContent string `json:"htmlContent"`
}

func SendEmailWithBrevo(toEmail, subject, html string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	senderEmail := os.Getenv("BREVO_SENDER_EMAIL")
	senderName := os.Getenv("BREVO_SENDER_NAME")

	if apiKey == "" || senderEmail == "" {
		return errors.New("Brevo env vars missing")
	}

	payload := brevoEmailRequest{}
	payload.Sender.Email = senderEmail
	payload.Sender.Name = senderName
	payload.Subject = subject
	payload.HTMLContent = html
	payload.To = append(payload.To, struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	}{
		Email: toEmail,
	})

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(
		"POST",
		"https://api.brevo.com/v3/smtp/email",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New("brevo email failed")
	}

	return nil
}
