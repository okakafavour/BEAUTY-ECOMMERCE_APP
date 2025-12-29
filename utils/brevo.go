package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	sib "github.com/sendinblue/APIv3-go-library/lib"
)

var brevoClient *sib.APIClient

// Initialize Brevo client
func InitBrevo() {
	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		log.Println("⚠️ BREVO_API_KEY not set")
		return
	}

	cfg := sib.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	brevoClient = sib.NewAPIClient(cfg)

	log.Println("✅ Brevo initialized")
}

// SendBrevoEmail sends an email via Brevo API
func SendBrevoEmail(toEmail, toName, subject, html string) error {
	if brevoClient == nil {
		return fmt.Errorf("Brevo client not initialized")
	}

	email := sib.SendSmtpEmail{
		Sender: &sib.SendSmtpEmailSender{
			Email: os.Getenv("BREVO_FROM"),
			Name:  os.Getenv("BREVO_FROM_NAME"),
		},
		To: []sib.SendSmtpEmailTo{
			{
				Email: toEmail,
				Name:  toName,
			},
		},
		Subject:     subject,
		HtmlContent: html,
	}

	_, _, err := brevoClient.TransactionalEmailsApi.SendTransacEmail(context.Background(), email)
	return err
}
