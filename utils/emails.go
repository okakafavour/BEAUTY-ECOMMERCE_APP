package utils

import (
	"fmt"
	"math"
	"net/smtp"
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

func SendEmail(to, subject, text, html string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	// ✅ Hard validation (VERY IMPORTANT)
	if host == "" || port == "" || username == "" || password == "" || from == "" {
		return fmt.Errorf("SMTP environment variables not fully configured")
	}

	addr := host + ":" + port

	auth := smtp.PlainAuth("", username, password, host)

	message := []byte(
		fmt.Sprintf("From: Beauty Shop <%s>\r\n", from) +
			fmt.Sprintf("To: %s\r\n", to) +
			fmt.Sprintf("Subject: %s\r\n", subject) +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
			html,
	)

	// ✅ Send with controlled failure
	if err := smtp.SendMail(addr, auth, from, []string{to}, message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func CreateStripePaymentIntentWithMetadata(
	amountGBP float64,
	orderID string,
	email string,
	name string,
) (*stripe.PaymentIntent, error) {

	amountInPence := int64(math.Round(amountGBP * 100))

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountInPence),
		Currency: stripe.String("gbp"),
	}

	// ✅ Correct way to attach metadata
	params.Metadata = map[string]string{
		"order_id":   orderID,
		"user_email": email,
		"user_name":  name,
	}

	return paymentintent.New(params)
}

// SendCustomerPaymentSuccess sends a confirmation email to the customer
func SendCustomerPaymentSuccess(userEmail, userName, orderID string) error {
	subject := fmt.Sprintf("Payment Successful ✅ - Order %s", orderID)
	text := fmt.Sprintf("Hello %s,\nYour payment for order %s was successful.\nThank you for shopping with Beauty Shop!", userName, orderID)
	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your payment for order <strong>%s</strong> was successful.</p>
		<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, userName, orderID)
	return SendEmail(userEmail, subject, text, html)
}

// SendCustomerPaymentFailed notifies the customer of failed payment
func SendCustomerPaymentFailed(userEmail, userName, orderID string) error {
	subject := fmt.Sprintf("Payment Failed ❌ - Order %s", orderID)
	text := fmt.Sprintf("Hello %s,\nYour payment for order %s has failed. Please try again or contact support.", userName, orderID)
	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your payment for order <strong>%s</strong> has failed.</p>
		<p>Please try again or contact support for assistance.</p>
	`, userName, orderID)
	return SendEmail(userEmail, subject, text, html)
}

// SendAdminNotification sends a notification to the admin
func SendAdminNotification(
	orderID,
	userName,
	userEmail,
	deliveryType string,
	subtotal,
	shippingFee,
	total float64,
) error {

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return nil // admin email not configured
	}

	subject := "New order placed on Beauty Shop"

	text := fmt.Sprintf(
		"A new order has been placed.\n\n"+
			"Customer name: %s\n"+
			"Customer email: %s\n"+
			"Order reference: %s\n"+
			"Delivery type: %s\n"+
			"Subtotal: %.2f GBP\n"+
			"Shipping fee: %.2f GBP\n"+
			"Order total: %.2f GBP\n",
		userName,
		userEmail,
		orderID,
		deliveryType,
		subtotal,
		shippingFee,
		total,
	)

	html := fmt.Sprintf(`
		<h2>New Order Notification</h2>

		<p>A new order has been placed on <strong>Beauty Shop</strong>.</p>

		<ul>
			<li><strong>Customer:</strong> %s</li>
			<li><strong>Email:</strong> %s</li>
			<li><strong>Order reference:</strong> %s</li>
			<li><strong>Delivery type:</strong> %s</li>
			<li><strong>Subtotal:</strong> %.2f GBP</li>
			<li><strong>Shipping fee:</strong> %.2f GBP</li>
			<li><strong>Order total:</strong> %.2f GBP</li>
		</ul>

		<hr />
		<p style="font-size:12px;color:#666;">
			Beauty Shop<br />
			Admin notification
		</p>
	`,
		userName,
		userEmail,
		orderID,
		deliveryType,
		subtotal,
		shippingFee,
		total,
	)

	return SendEmail(adminEmail, subject, text, html)
}
