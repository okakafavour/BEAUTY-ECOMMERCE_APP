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

	// Queue welcome email via Brevo
	subject := "Welcome to Beauty Shop ✨"
	html := fmt.Sprintf(`
	<h2>Hello %s 👋</h2>
	<p>Your account has been created successfully.</p>
	<p>Welcome to Beauty Shop 💄</p>
`, user.Name)
	utils.QueueEmail(user.Email, user.Name, subject, html)

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
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

// ForgotPassword sends password reset email
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
		frontendURL = "http://localhost:3000"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	// Queue reset email via Brevo
	utils.SendResetPasswordEmail(user.Email, user.Name, resetLink)

	c.JSON(200, gin.H{"message": "If email exists, reset link sent"})
}

// ResetPassword handles GET (link click) and POST (update password) requests
func ResetPassword(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		tokenQuery := c.Query("token")
		if tokenQuery == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenQuery})
		return
	}

	type Request struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
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

	hashedPassword, _ := utils.HashPassword(req.NewPassword)
	_ = userService.UpdatePassword(user.ID, hashedPassword)
	_ = userService.ClearResetToken(user.ID)

	// Notify user asynchronously via Brevo
	subject := "Your password has been reset"
	html := fmt.Sprintf(`
	<h2>Hello %s,</h2>
	<p>Your password has been successfully reset.</p>
	<p>If this wasn't you, please contact support immediately.</p>`, user.Name)
	utils.QueueEmail(user.Email, user.Name, subject, html)

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}
