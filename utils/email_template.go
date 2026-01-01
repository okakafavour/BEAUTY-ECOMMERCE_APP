package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
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
var brevoAPIKey string

// -----------------------------
// Initialize Brevo
// -----------------------------
func InitBrevo() {
	brevoAPIKey = os.Getenv("BREVO_API_KEY")
	if brevoAPIKey == "" {
		log.Println("⚠️ BREVO_API_KEY not set")
		return
	}
	log.Println("✅ Brevo initialized")
}

// -----------------------------
// Send email via Brevo with retries and timeout
// -----------------------------
func sendBrevoEmail(toEmail, toName, subject, html string) error {
	if brevoAPIKey == "" {
		return fmt.Errorf("Brevo API key not set")
	}

	fromEmail := os.Getenv("BREVO_SENDER_EMAIL")
	fromName := os.Getenv("BREVO_SENDER_NAME")

	client := resty.New().
		SetTimeout(30 * time.Second) // ← TLS/network timeout

	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  fromName,
			"email": fromEmail,
		},
		"to": []map[string]string{
			{
				"email": toEmail,
				"name":  toName,
			},
		},
		"subject":     subject,
		"htmlContent": html,
	}

	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.R().
			SetHeader("api-key", brevoAPIKey).
			SetHeader("Content-Type", "application/json").
			SetBody(payload).
			Post("https://api.brevo.com/v3/smtp/email")

		if err != nil {
			lastErr = err
			log.Printf("⚠️ Attempt %d: Failed to send email to %s: %v\n", attempt, toEmail, err)
			time.Sleep(time.Duration(attempt) * time.Second) // incremental backoff
			continue
		}

		if resp.StatusCode() >= 400 {
			lastErr = fmt.Errorf("Brevo error: %d - %s", resp.StatusCode(), resp.String())
			log.Printf("⚠️ Attempt %d: Brevo returned error for %s: %v\n", attempt, toEmail, lastErr)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		// Success
		log.Println("✅ Email sent:", toEmail)
		return nil
	}

	return fmt.Errorf("failed to send email to %s after %d attempts: last error: %v", toEmail, maxRetries, lastErr)
}

// -----------------------------
// Start Email Worker
// -----------------------------
func StartEmailWorker() {
	go func() {
		for job := range EmailQueue {
			err := sendBrevoEmail(job.To, job.ToName, job.Subject, job.HTML)
			if err != nil {
				log.Println("❌ Email permanently failed:", job.To, err)
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
