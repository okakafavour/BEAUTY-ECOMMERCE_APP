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
