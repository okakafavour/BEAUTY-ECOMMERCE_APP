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

	if user.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if err := userService.Register(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject := "Welcome to Beauty Shop!"
	plainText := "Hello " + user.Name + ", welcome to Beauty Shop! We're excited to have you on board."
	htmlContent := "<h1>Hello " + user.Name + "</h1><p>Welcome to Beauty Shop! We're excited to have you on board.</p>"

	// Send email asynchronously
	go func() {
		if err := utils.SendEmail(user.Email, subject, plainText, htmlContent); err != nil {
			fmt.Println("Failed to send welcome email:", err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
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
