package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
	servicesimpl "beauty-ecommerce-backend/services_impl"
	"beauty-ecommerce-backend/utils"
	"fmt"
	"net/http"

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

	// Prepare email content
	subject := "Welcome to Beauty Shop!"
	plainText := "Hello " + user.Name + ", welcome to Beauty Shop! We're excited to have you on board."
	htmlContent := "<h1>Hello " + user.Name + "</h1><p>Welcome to Beauty Shop! We're excited to have you on board.</p>"

	// Send email and log error synchronously for now
	err := utils.SendEmail(user.Email, subject, plainText, htmlContent)
	if err != nil {
		fmt.Println("Failed to send welcome email:", err)
		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully, but failed to send email", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully, email sent"})
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
