package utils

import (
	"fmt"
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

func CreateStripePaymentIntentWithMetadata(amount float64, currency, orderID, userEmail, userName string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(int64(amount * 100)),
		Currency:           stripe.String(currency),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	}

	// Add metadata
	params.AddMetadata("order_id", orderID)
	params.AddMetadata("user_email", userEmail)
	params.AddMetadata("user_name", userName)

	return paymentintent.New(params)
}
