package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	sib "github.com/sendinblue/APIv3-go-library/lib"
)

// -----------------------------
// Email Queue
// -----------------------------
type EmailJob struct {
	To      string
	ToName  string
	Subject string
	HTML    string
}

var EmailQueue = make(chan EmailJob, 100)
var brevoClient *sib.APIClient

// -----------------------------
// Initialize Brevo
// -----------------------------
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

// -----------------------------
// Send email via Brevo
// -----------------------------
func sendBrevoEmail(toEmail, toName, subject, html string) error {
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

// -----------------------------
// Start Email Worker
// -----------------------------
func StartEmailWorker() {
	go func() {
		for job := range EmailQueue {
			err := sendBrevoEmail(
				job.To,
				job.ToName,
				job.Subject,
				job.HTML,
			)
			if err != nil {
				log.Println("⚠️ Email failed:", job.To, err)
			} else {
				log.Println("✅ Email sent:", job.To)
			}
		}
	}()
}

// -----------------------------
// Queue Email (non-blocking)
// -----------------------------
func QueueEmail(to, toName, subject, html string) {
	select {
	case EmailQueue <- EmailJob{
		To:      to,
		ToName:  toName,
		Subject: subject,
		HTML:    html,
	}:
	default:
		log.Println("⚠️ Email queue full, skipping email to", to)
	}
	log.Println("📩 Queue email to:", to)
}

// -----------------------------
// Email Templates
// -----------------------------
func SendConfirmationEmail(to, name, orderID, deliveryType string, subtotal, shippingFee, total float64) {
	subject := "Your order payment update"
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>We’ve received your payment and your order is now being processed.</p>
	<p><strong>Order reference:</strong> %s</p>
	<ul>
		<li>Delivery type: %s</li>
		<li>Subtotal: %.2f GBP</li>
		<li>Shipping fee: %.2f GBP</li>
		<li><strong>Order total: %.2f GBP</strong></li>
	</ul>
	<p>We’ll notify you once your order is shipped.</p>
	<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, name, orderID, deliveryType, subtotal, shippingFee, total)

	QueueEmail(to, name, subject, html)
}

func SendFailedPaymentEmail(to, name, orderID string) {
	subject := "Issue with your order payment"
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>There was an issue processing your payment.</p>
	<p><strong>Order reference:</strong> %s</p>
	<p>Please try again or contact support.</p>
	`, name, orderID)

	QueueEmail(to, name, subject, html)
}

func SendResetPasswordEmail(toEmail, name, resetLink string) {
	subject := "Reset your Beauty Shop password"
	html := fmt.Sprintf(`
	<h2>Password Reset</h2>
	<p>Click the button below to reset your password:</p>
	<p><a href="%s">Reset Password</a></p>
	<p>This link expires in 15 minutes.</p>
	`, resetLink)

	QueueEmail(toEmail, name, subject, html)
}

func SendShipmentEmail(toEmail, toName, orderID, deliveryType string) {
	subject := "Your Order Has Been Shipped!"
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>Your order <b>%s</b> has been shipped.</p>
	<p>Delivery type: %s</p>
	<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, toName, orderID, deliveryType)

	QueueEmail(toEmail, toName, subject, html)
}
