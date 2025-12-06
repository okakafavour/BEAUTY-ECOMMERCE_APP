package utils

import "fmt"

func OrderConfirmationEmail(name, orderId string) (string, string) {
	subject := "Order Confirmation - " + orderId

	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your order <b>%s</b> has been received successfully!</p>
		<p>We will notify you once it's shipped.</p>
		<br>
		<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, name, orderId)

	return subject, html
}

func SendConfirmationEmail(to, name, orderID string) error {
	subject := fmt.Sprintf("Payment Successful ✅ - Order %s", orderID)

	text := fmt.Sprintf("Hello %s,\nYour payment for order %s was successful.\nThank you for shopping with Beauty Shop!", name, orderID)
	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your payment for order <strong>%s</strong> was successful.</p>
		<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, name, orderID)

	return SendEmail(to, subject, text, html)
}

func SendFailedPaymentEmail(to, name, orderID string) error {
	subject := fmt.Sprintf("Payment Failed ❌ - Order %s", orderID)

	text := fmt.Sprintf("Hello %s,\nYour payment for order %s has failed. Please try again or contact support.", name, orderID)
	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your payment for order <strong>%s</strong> has failed.</p>
		<p>Please try again or contact support for assistance.</p>
	`, name, orderID)

	return SendEmail(to, subject, text, html)
}
