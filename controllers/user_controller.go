package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
	servicesimpl "beauty-ecommerce-backend/services_impl"
	"beauty-ecommerce-backend/utils"
	"fmt"
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

	// ✅ Respond immediately (DO NOT BLOCK)
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
	})

	// ✅ Prepare email content
	subject := "Welcome to Beauty Shop 🎉"
	plainText := fmt.Sprintf(
		"Hello %s,\n\nWelcome to Beauty Shop! We're excited to have you on board.",
		user.Name,
	)

	htmlContent := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Welcome to <strong>Beauty Shop</strong>! 🎉</p>
		<p>We’re excited to have you on board.</p>
	`, user.Name)

	// ✅ Send email asynchronously (SAFE)
	go func() {
		if err := utils.SendEmail(user.Email, subject, plainText, htmlContent); err != nil {
			fmt.Println("Email error:", err)
		}
	}()
}

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

// Test endpoint to confirm email sending works
func TestEmail(c *gin.Context) {
	// Replace with your email to test
	err := utils.SendEmail("testuser@gmail.com", "Test Email", "This is a test", "<p>This is a test</p>")
	if err != nil {
		fmt.Println("Error sending email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Email sent successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Email sent"})
}

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
		// Don't reveal if email exists
		c.JSON(200, gin.H{"message": "If email exists, reset link sent"})
		return
	}

	// Generate token
	token := utils.GenerateRandomToken(32)

	// Hash token before saving
	hashedToken := utils.HashToken(token)

	expiry := time.Now().Add(15 * time.Minute)

	err = userService.SavePasswordResetToken(
		user.ID,
		hashedToken,
		expiry,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not save token"})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8080" // fallback for local testing
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	go utils.SendResetPasswordEmail(user.Email, resetLink)

	c.JSON(200, gin.H{
		"message": "If email exists, reset link sent",
	})
}

func ResetPassword(c *gin.Context) {
	// -----------------------------
	// Handle GET request (link click)
	// -----------------------------
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

	// -----------------------------
	// Handle POST request (reset password)
	// -----------------------------
	type Request struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input, password must be at least 8 characters"})
		return
	}

	// Hash token to compare with DB
	hashedToken := utils.HashToken(req.Token)

	// Get user by reset token
	user, err := userService.GetUserByResetToken(hashedToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	// Check token expiry
	if time.Now().After(user.ResetPasswordExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token expired"})
		return
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Update password in DB
	if err := userService.UpdatePassword(user.ID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Clear reset token so old password cannot be reused
	_ = userService.ClearResetToken(user.ID)

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}
