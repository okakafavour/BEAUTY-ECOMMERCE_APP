package utils

import (
	"fmt"
	"math"
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

// CreateStripePaymentIntentWithMetadata creates a Stripe PaymentIntent with metadata
func CreateStripePaymentIntentWithMetadata(
	amountGBP float64,
	orderID string,
	email string,
	name string,
	deliveryType string,
	subtotal, shippingFee, total float64,
) (*stripe.PaymentIntent, error) {

	amountInPence := int64(math.Round(amountGBP * 100))

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountInPence),
		Currency: stripe.String("gbp"),
	}

	// Attach full order metadata
	params.Metadata = map[string]string{
		"order_id":      orderID,
		"user_email":    email,
		"user_name":     name,
		"delivery_type": deliveryType,
		"subtotal":      fmt.Sprintf("%.2f", subtotal),
		"shipping_fee":  fmt.Sprintf("%.2f", shippingFee),
		"total_price":   fmt.Sprintf("%.2f", total),
	}

	return paymentintent.New(params)
}

// SendCustomerPaymentSuccess queues a confirmation email to the customer
func SendCustomerPaymentSuccess(userEmail, userName, orderID, deliveryType string, subtotal, shippingFee, total float64) {
	subject := "Payment Successful ✅ - Order " + orderID
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>Your payment for order <strong>%s</strong> was successful.</p>
	<p>Delivery type: %s</p>
	<p>Subtotal: %.2f GBP | Shipping: %.2f GBP | Total: %.2f GBP</p>
	<p>Thank you for shopping with Beauty Shop ❤️</p>
	<hr />
	<p style="font-size:12px;color:#666;">Beauty Shop<br/>Official payment notification<br/>Support: support@batluxebeauty.com</p>
	`, userName, orderID, deliveryType, subtotal, shippingFee, total)

	QueueEmail(userEmail, subject, "", html)
}

// SendAdminNotification queues an email to the admin
func SendAdminNotification(
	orderID, userName, userEmail, deliveryType string,
	subtotal, shippingFee, total float64,
) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return // admin email not configured
	}

	subject := "New Order Paid - " + orderID
	html := fmt.Sprintf(`
	<h2>New Order Notification</h2>
	<p>A new order has been successfully paid on <strong>Beauty Shop</strong>.</p>
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
	<p style="font-size:12px;color:#666;">Beauty Shop<br/>Admin notification</p>
	`, userName, userEmail, orderID, deliveryType, subtotal, shippingFee, total)

	QueueEmail(adminEmail, subject, "", html)
}
