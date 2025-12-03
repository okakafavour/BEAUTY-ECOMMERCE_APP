package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func SendEmail(to, subject, text, html string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	senderEmail := os.Getenv("BREVO_SENDER_EMAIL")
	senderName := os.Getenv("BREVO_SENDER_NAME")

	body := map[string]interface{}{
		"sender": map[string]string{
			"name":  senderName,
			"email": senderEmail,
		},
		"to": []map[string]string{
			{"email": to},
		},
		"subject":     subject,
		"textContent": text,
		"htmlContent": html,
	}

	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonBody))
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
		return fmt.Errorf("failed: %s", resp.Status)
	}
	return nil
}
