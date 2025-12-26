package utils

import "fmt"

func OrderConfirmationEmail(name, orderID, deliveryType string, subtotal, shippingFee, total float64) (string, string) {
	subject := "We’ve received your order"

	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>

		<p>We’ve received your order and it is currently being prepared.</p>

		<p><strong>Order reference:</strong> %s</p>

		<ul>
			<li>Delivery type: %s</li>
			<li>Subtotal: %.2f GBP</li>
			<li>Shipping fee: %.2f GBP</li>
			<li><strong>Order total: %.2f GBP</strong></li>
		</ul>

		<p>We’ll notify you as soon as your order is shipped.</p>

		<p>Thank you for shopping with Beauty Shop ❤️</p>

		<hr />
		<p style="font-size:12px;color:#666;">
			Beauty Shop<br />
			Official order notification<br />
			Support: support@batluxebeauty.com
		</p>
	`, name, orderID, deliveryType, subtotal, shippingFee, total)

	return subject, html
}

func SendConfirmationEmail(to, name, orderID, deliveryType string, subtotal, shippingFee, total float64) error {
	subject := "Your order payment update"

	text := fmt.Sprintf(
		"Hello %s,\n\n"+
			"We’ve received your payment and your order is being processed.\n\n"+
			"Order reference: %s\n"+
			"Delivery type: %s\n"+
			"Subtotal: %.2f GBP\n"+
			"Shipping fee: %.2f GBP\n"+
			"Order total: %.2f GBP\n\n"+
			"We’ll notify you when your order is shipped.\n\n"+
			"Beauty Shop\nSupport: support@batluxebeauty.com",
		name, orderID, deliveryType, subtotal, shippingFee, total,
	)

	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>

		<p>We’ve received your payment and your order is now being processed.</p>

		<p><strong>Order reference:</strong> %s</p>

		<ul>
			<li>Delivery type: %s</li>
			<li>Subtotal: %.2f GBP</li>
			<li>Shipping fee: %.2f GBP</li>
			<li><strong>Order total: %.2f GBP</strong></li>
		</ul>

		<p>We’ll notify you once your order is shipped.</p>

		<p>Thank you for shopping with Beauty Shop ❤️</p>

		<hr />
		<p style="font-size:12px;color:#666;">
			Beauty Shop<br />
			Official payment notification<br />
			Support: support@batluxebeauty.com
		</p>
	`, name, orderID, deliveryType, subtotal, shippingFee, total)

	return SendEmail(to, subject, text, html)
}

func SendFailedPaymentEmail(to, name, orderID string) error {
	subject := "Issue with your order payment"

	text := fmt.Sprintf(
		"Hello %s,\n\n"+
			"There was an issue processing your payment for the order below:\n\n"+
			"Order reference: %s\n\n"+
			"Please try again or contact our support team if you need help.\n\n"+
			"Beauty Shop\nSupport: support@batluxebeauty.com",
		name, orderID,
	)

	html := fmt.Sprintf(`
		<h2>Hello %s,</h2>

		<p>There was an issue processing your payment.</p>

		<p><strong>Order reference:</strong> %s</p>

		<p>Please try again or contact our support team for assistance.</p>

		<hr />
		<p style="font-size:12px;color:#666;">
			Beauty Shop<br />
			Support: support@batluxebeauty.com
		</p>
	`, name, orderID)

	return SendEmail(to, subject, text, html)
}

func SendResetPasswordEmail(toEmail, resetLink string) error {
	subject := "Reset your Beauty Shop password"

	plainText := fmt.Sprintf(
		"Click the link below to reset your password:\n\n%s\n\n"+
			"This link expires in 15 minutes.\n\n"+
			"If you didn’t request this, you can safely ignore this email.\n\n"+
			"Beauty Shop Support",
		resetLink,
	)

	html := fmt.Sprintf(`
		<h2>Password Reset</h2>

		<p>Click the button below to reset your password.</p>

		<p>
			<a href="%s" style="padding:10px 15px;background:#000;color:#fff;text-decoration:none;">
				Reset Password
			</a>
		</p>

		<p>This link expires in 15 minutes.</p>

		<p>If you didn’t request this, you can safely ignore this email.</p>

		<hr />
		<p style="font-size:12px;color:#666;">
			Beauty Shop<br />
			Security notification
		</p>
	`, resetLink)

	return SendEmail(toEmail, subject, plainText, html)
}
