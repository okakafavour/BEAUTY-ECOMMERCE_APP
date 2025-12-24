package utils

import "fmt"

func OrderConfirmationEmail(name, orderID, deliveryType string, subtotal, shippingFee, total float64) (string, string) {
	subject := "Order Confirmation - " + orderID

	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your order <b>%s</b> has been received successfully!</p>
		<ul>
			<li>Delivery Type: %s</li>
			<li>Subtotal: £%.2f</li>
			<li>Shipping Fee: £%.2f</li>
			<li><strong>Total: £%.2f</strong></li>
		</ul>
		<p>We will notify you once it's shipped.</p>
		<br>
		<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, name, orderID, deliveryType, subtotal, shippingFee, total)

	return subject, html
}

func SendConfirmationEmail(to, name, orderID, deliveryType string, subtotal, shippingFee, total float64) error {
	subject := fmt.Sprintf("Payment Successful ✅ - Order %s", orderID)
	text := fmt.Sprintf("Hello %s,\nYour payment for order %s was successful.\nDelivery: %s\nSubtotal: £%.2f\nShipping: £%.2f\nTotal: £%.2f\nThank you for shopping with Beauty Shop!", name, orderID, deliveryType, subtotal, shippingFee, total)

	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your payment for order <strong>%s</strong> was successful.</p>
		<ul>
			<li>Delivery Type: %s</li>
			<li>Subtotal: £%.2f</li>
			<li>Shipping Fee: £%.2f</li>
			<li><strong>Total: £%.2f</strong></li>
		</ul>
		<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, name, orderID, deliveryType, subtotal, shippingFee, total)

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

func SendResetPasswordEmail(toEmail, resetLink string) error {
	subject := "Reset your password"
	plainText := fmt.Sprintf(
		"Click the link below to reset your password:\n\n%s\n\nThis link expires in 15 minutes.",
		resetLink,
	)

	html := fmt.Sprintf(`
		<h2>Password Reset</h2>
		<p>Click the button below to reset your password.</p>
		<p><a href="%s" style="padding:10px 15px;background:#000;color:#fff;text-decoration:none;">
		Reset Password</a></p>
		<p>This link expires in 15 minutes.</p>
	`, resetLink)

	return SendEmail(toEmail, subject, plainText, html)
}
