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
