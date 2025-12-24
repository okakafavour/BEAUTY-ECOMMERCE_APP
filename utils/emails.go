package utils

import (
	"fmt"
	"math"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

func SendEmail(to, subject, text, html string) error {
	from := mail.NewEmail("Beauty Shop", os.Getenv("EMAIL_FROM")) // Verified sender
	toEmail := mail.NewEmail("", to)

	message := mail.NewSingleEmail(from, subject, toEmail, text, html)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
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
func SendAdminNotification(eventType, orderID, userName, userEmail string) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return nil // no admin configured
	}

	subject := fmt.Sprintf("%s - Order %s", eventType, orderID)
	text := fmt.Sprintf("Order %s %s by %s (%s)", orderID, eventType, userName, userEmail)
	html := fmt.Sprintf("<p>Order <b>%s</b> %s by <b>%s</b> (%s).</p>", orderID, eventType, userName, userEmail)

	return SendEmail(adminEmail, subject, text, html)
}
