package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
	servicesimpl "beauty-ecommerce-backend/services_impl"
	"beauty-ecommerce-backend/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userService services.UserService

func InitUserController(userRepo *repositories.UserRepository) {
	userService = servicesimpl.NewUserService(userRepo)
}

// Register creates a new user and sends welcome email via Brevo
func Register(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	if err := userService.Register(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("🧪 Register: user created, sending welcome email")

	// Send welcome email asynchronously via MailerSend
	go func(email, name string) {
		subject := "Welcome to Beauty Shop ✨"
		html := fmt.Sprintf(`
		<h2>Hello %s 👋</h2>
		<p>Your account has been created successfully.</p>
		<p>Welcome to Beauty Shop 💄</p>
	`, name)

		if err := utils.SendMailSenderEmail(email, name, subject, html); err != nil {
			log.Println("⚠️ Welcome email failed:", err)
		} else {
			log.Println("✅ Welcome email sent to", email)
		}
	}(user.Email, user.Name)

}

// Login authenticates user and returns JWT token
func Login(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	token, err := userService.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
}

// TestEmail sends a test email via Brevo
func TestEmail(c *gin.Context) {
	go func() {
		err := utils.SendEmailWithBrevo("testuser@gmail.com", "Test Email", "<p>This is a test email from Brevo</p>")
		if err != nil {
			fmt.Println("⚠️ Error sending test email:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("✅ Test email sent successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Email sent"})
	}()
}

// GetProfile returns the current user profile
func GetProfile(c *gin.Context) {
	userID, _ := utils.ExtractUserIDAndRole(c)

	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// ForgotPassword generates a reset token and sends it via Brevo
func ForgotPassword(c *gin.Context) {
	type Request struct {
		Email string `json:"email" binding:"required,email"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid email"})
		return
	}

	user, err := userService.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(200, gin.H{"message": "If email exists, reset link sent"})
		return
	}

	token := utils.GenerateRandomToken(32)
	hashedToken := utils.HashToken(token)
	expiry := time.Now().Add(15 * time.Minute)

	err = userService.SavePasswordResetToken(user.ID, hashedToken, expiry)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not save token"})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8080"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	// Send reset password email asynchronously via Brevo
	go func() {
		subject := "Password Reset Request"
		html := fmt.Sprintf("<p>Click to reset your password: <a href='%s'>Reset Password</a></p>", resetLink)
		if err := utils.SendEmailWithBrevo(user.Email, subject, html); err != nil {
			fmt.Println("⚠️ Failed to send reset password email:", err)
		} else {
			fmt.Println("✅ Reset password email sent to", user.Email)
		}
	}()

	c.JSON(200, gin.H{
		"message": "If email exists, reset link sent",
	})
}

// ResetPassword handles GET (link click) and POST (update password) requests
func ResetPassword(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		tokenQuery := c.Query("token")
		if tokenQuery == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token":   tokenQuery,
			"message": "Token received. Use POST to reset password with new password.",
		})
		return
	}

	type Request struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input, password must be at least 8 characters"})
		return
	}

	hashedToken := utils.HashToken(req.Token)
	user, err := userService.GetUserByResetToken(hashedToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	if time.Now().After(user.ResetPasswordExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token expired"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := userService.UpdatePassword(user.ID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	_ = userService.ClearResetToken(user.ID)

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}
