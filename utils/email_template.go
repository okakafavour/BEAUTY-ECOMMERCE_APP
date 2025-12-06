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

func SendConfirmationEmail(toEmail, userName, orderID string) error {
	subject := "Payment Successful ✅"
	body := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>Your payment for order <strong>%s</strong> was successful.</p>
	<p>Thank you for shopping with us!</p>
	`, userName, orderID)

	// TODO: Replace this with real email sending code
	fmt.Printf("Sending email to %s\nSubject: %s\nBody: %s\n", toEmail, subject, body)

	// Return nil for now since we are just printing
	return nil
}
func SendFailedPaymentEmail(to string, name string, orderID string) error {
	fmt.Println("⚠️ Sending failed payment email to:", to)
	fmt.Println("Name:", name)
	fmt.Println("OrderID:", orderID)
	return nil
}
