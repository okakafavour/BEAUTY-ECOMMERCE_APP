package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var orderService services.OrderService

func InitOrderController(os services.OrderService) {
	orderService = os
}

func CreateOrder(c *gin.Context) {
	var order models.Order

	// Bind JSON
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT middleware
	rawUserID, _ := c.Get("user_id")
	userIDString := rawUserID.(string)

	// Convert to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}

	// Assign
	order.UserID = userID

	// Call service
	createdOrder, err := orderService.CreateOrder(order)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, createdOrder)
}

// GET /orders (get orders for logged in user)
func GetOrders(c *gin.Context) {
	userIDHex, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := primitive.ObjectIDFromHex(userIDHex.(string))

	orders, err := orderService.GetOrdersByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GET /orders/:id
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
