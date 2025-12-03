package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// -------------------------------
// Shared package-level OrderService
// -------------------------------
var orderService services.OrderService

// Initialize the service once at app startup
func InitOrderController(os services.OrderService) {
	orderService = os
}

// -------------------------------
// Create Order
// POST /orders
// -------------------------------
func CreateOrder(c *gin.Context) {
	var order models.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := rawUserID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	order.UserID = userID

	createdOrder, err := orderService.CreateOrder(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send response first
	c.JSON(http.StatusCreated, createdOrder)

	// Send order confirmation email asynchronously
	go func() {
		// Fetch user details from UserService
		user, err := userService.GetUserByID(userID) // <- You need to implement this in UserService
		if err != nil {
			fmt.Println("Failed to fetch user for email:", err)
			return
		}

		subject, html := utils.OrderConfirmationEmail(user.Name, createdOrder.ID.Hex())
		err = utils.SendEmail(user.Email, subject, "", html)
		if err != nil {
			fmt.Println("Failed to send order confirmation email:", err)
		}
	}()
}

// -------------------------------
// Get Orders for logged-in user
// GET /orders
// -------------------------------
func GetOrders(c *gin.Context) {
	rawUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := primitive.ObjectIDFromHex(rawUserID.(string))

	orders, err := orderService.GetOrdersByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// -------------------------------
// Get single order
// GET /orders/:id
// -------------------------------
func GetOrderByID(c *gin.Context) {
	orderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := orderService.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// -------------------------------
// Cancel order
// PUT /orders/:id/cancel
// -------------------------------
func CancelOrder(c *gin.Context) {
	orderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	rawUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(rawUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	updatedOrder, err := orderService.CancelOrder(orderID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"order":   updatedOrder,
	})
}
