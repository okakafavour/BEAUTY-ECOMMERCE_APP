package utils

// import (
// 	"fmt"
// 	"os"
// )

// // SendTestAdminEmail sends a test email to the admin to verify SMTP
// func SendTestAdminEmail() error {
// 	adminEmail := os.Getenv("ADMIN_EMAIL")
// 	if adminEmail == "" {
// 		return fmt.Errorf("ADMIN_EMAIL not set in .env")
// 	}

// 	subject := "Test Admin Email ✅"
// 	text := "This is a test email to check if admin receives messages."
// 	html := "<p>This is a <strong>test email</strong> to admin.</p>"

// 	err := SendEmail(adminEmail, subject, text, html)
// 	if err != nil {
// 		return fmt.Errorf("failed to send admin email: %v", err)
// 	}

// 	fmt.Println("Admin test email sent successfully!")
// 	return nil
// }
