package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type BrevoEmail struct {
	Sender      map[string]string   `json:"sender"`
	To          []map[string]string `json:"to"`
	Subject     string              `json:"subject"`
	HtmlContent string              `json:"htmlContent"`
}

func SendEmailWithBrevo(toEmail, subject, htmlContent string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	senderEmail := os.Getenv("BREVO_SENDER_EMAIL")
	senderName := os.Getenv("BREVO_SENDER_NAME")

	email := BrevoEmail{
		Sender: map[string]string{
			"email": senderEmail,
			"name":  senderName,
		},
		To: []map[string]string{
			{"email": toEmail},
		},
		Subject:     subject,
		HtmlContent: htmlContent,
	}

	payload, err := json.Marshal(email)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("brevo API returned status %d", resp.StatusCode)
	}

	return nil
}
