package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mailersend/mailersend-go"
)

var ms *mailersend.Mailersend

func InitMailerSend() {
	apiKey := os.Getenv("MAILERSEND_API_KEY")
	if apiKey == "" {
		log.Println("⚠️ MAILERSEND_API_KEY not set")
		return
	}

	ms = mailersend.NewMailersend(apiKey)
	log.Println("✅ MailerSend initialized")
}

func SendEmail(toEmail, toName, subject, html string) error {
	if ms == nil {
		return fmt.Errorf("MailerSend not initialized")
	}

	message := ms.Email.NewMessage()

	message.SetFrom(mailersend.Recipient{
		Email: os.Getenv("MAILERSEND_FROM"),
		Name:  os.Getenv("MAILERSEND_FROM_NAME"),
	})

	message.SetRecipients([]mailersend.Recipient{
		{
			Email: toEmail,
			Name:  toName,
		},
	})

	message.SetSubject(subject)
	message.SetHTML(html)

	_, err := ms.Email.Send(context.Background(), message)
	if err != nil {
		return err
	}

	return nil
}
