package utils

import (
	"fmt"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

// CreateStripePaymentIntent creates a PaymentIntent for a given order
func CreateStripePaymentIntent(amount float64, currency, orderID, userEmail string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)), // Convert to kobo
		Currency: stripe.String(currency),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
	}

	// Add metadata
	params.AddMetadata("order_id", orderID)
	params.AddMetadata("user_email", userEmail)

	return paymentintent.New(params)
}

func PaymentSuccessEmail(userName, orderID string) (subject string, htmlBody string) {
	subject = "Payment Successful ✅"
	htmlBody = fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your payment for order <strong>%s</strong> was successful.</p>
		<p>Thank you for shopping with us!</p>
	`, userName, orderID)
	return
}
