package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// -----------------------------
// Email Job Queue (optional for scale)
// -----------------------------
type EmailJob struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

var EmailQueue = make(chan EmailJob, 100) // buffered channel

// StartEmailWorker starts a goroutine to process email jobs
func StartEmailWorker() {
	go func() {
		for job := range EmailQueue {
			if err := sendSMTP(job.To, job.Subject, job.Text, job.HTML); err != nil {
				log.Println("⚠️ Failed to send email to", job.To, ":", err)
			} else {
				log.Println("✅ Email sent to:", job.To)
			}
		}
	}()
}

// QueueEmail adds a new email to the queue (non-blocking)
func QueueEmail(to, subject, text, html string) {
	EmailQueue <- EmailJob{To: to, Subject: subject, Text: text, HTML: html}
}

// -----------------------------
// Core SMTP function
// -----------------------------
func sendSMTP(to, subject, text, html string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if from == "" || password == "" || host == "" || port == "" {
		return fmt.Errorf("SMTP configuration is missing in .env")
	}

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		html

	auth := smtp.PlainAuth("", from, password, host)
	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))
}

// -----------------------------
// Utility: order emails
// -----------------------------
func OrderConfirmationEmail(name, orderID, deliveryType string, subtotal, shippingFee, total float64) (string, string) {
	subject := "We’ve received your order"
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>We’ve received your order and it is currently being prepared.</p>
	<p><strong>Order reference:</strong> %s</p>
	<ul>
		<li>Delivery type: %s</li>
		<li>Subtotal: %.2f GBP</li>
		<li>Shipping fee: %.2f GBP</li>
		<li><strong>Order total: %.2f GBP</strong></li>
	</ul>
	<p>We’ll notify you as soon as your order is shipped.</p>
	<p>Thank you for shopping with Beauty Shop ❤️</p>
	<hr />
	<p style="font-size:12px;color:#666;">Beauty Shop<br/>Official order notification<br/>Support: support@batluxebeauty.com</p>
	`, name, orderID, deliveryType, subtotal, shippingFee, total)
	return subject, html
}

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
	<hr />
	<p style="font-size:12px;color:#666;">Beauty Shop<br/>Official payment notification<br/>Support: support@batluxebeauty.com</p>
	`, name, orderID, deliveryType, subtotal, shippingFee, total)

	// Queue email (non-blocking)
	QueueEmail(to, subject, "", html)
}

func SendFailedPaymentEmail(to, name, orderID string) {
	subject := "Issue with your order payment"
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>There was an issue processing your payment.</p>
	<p><strong>Order reference:</strong> %s</p>
	<p>Please try again or contact our support team for assistance.</p>
	<hr />
	<p style="font-size:12px;color:#666;">Beauty Shop<br/>Support: support@batluxebeauty.com</p>
	`, name, orderID)

	QueueEmail(to, subject, "", html)
}

func SendResetPasswordEmail(toEmail, resetLink string) {
	subject := "Reset your Beauty Shop password"
	html := fmt.Sprintf(`
	<h2>Password Reset</h2>
	<p>Click the button below to reset your password:</p>
	<p><a href="%s" style="padding:10px 15px;background:#000;color:#fff;text-decoration:none;">Reset Password</a></p>
	<p>This link expires in 15 minutes.</p>
	<p>If you didn’t request this, you can safely ignore this email.</p>
	<hr />
	<p style="font-size:12px;color:#666;">Beauty Shop<br/>Security notification</p>
	`, resetLink)

	QueueEmail(toEmail, subject, "", html)
}

func SendShipmentEmail(toEmail, toName, orderID, deliveryType string) {
	subject := "Your Order Has Been Shipped!"
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>Good news! Your order <b>%s</b> has been shipped.</p>
	<p>Delivery type: %s</p>
	<p>You can expect it to arrive soon.</p>
	<p>Thank you for shopping with Beauty Shop ❤️</p>
	<hr />
	<p style="font-size:12px;color:#666;">Beauty Shop<br/>Official order notification<br/>Support: support@batluxebeauty.com</p>
	`, toName, orderID, deliveryType)

	QueueEmail(toEmail, subject, "", html)
}
