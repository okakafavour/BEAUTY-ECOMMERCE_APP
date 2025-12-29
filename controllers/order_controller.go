package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var orderService services.OrderService

func InitOrderController(os services.OrderService, us services.UserService) {
	orderService = os
	userService = us
}

// CreateOrder handles POST /orders
func CreateOrder(c *gin.Context) {
	var order models.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.DeliveryType != "standard" && order.DeliveryType != "express" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery type"})
		return
	}

	switch order.DeliveryType {
	case "standard":
		order.ShippingFee = 3.99
	case "express":
		order.ShippingFee = 4.99
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

	var subtotal float64
	for i, item := range order.Items {
		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		product, err := orderService.GetProductByID(productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price
		subtotal += product.Price * float64(item.Quantity)
	}

	order.Subtotal = subtotal
	order.TotalPrice = subtotal + order.ShippingFee
	order.Status = "pending"

	createdOrder, err := orderService.CreateOrder(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdOrder)

	// Fetch user info
	user, err := userService.GetUserByID(userID)
	if err != nil {
		log.Println("⚠️ Failed to fetch user for order email:", err)
		return
	}

	// Queue order confirmation email via Brevo
	subject := fmt.Sprintf("Order Confirmation - %s", createdOrder.ID.Hex())
	html := fmt.Sprintf(`
	<h2>Hello %s 👋</h2>
	<p>Your order <b>%s</b> has been received successfully!</p>
	<ul>
		<li>Delivery Type: %s</li>
		<li>Subtotal: $%.2f</li>
		<li>Shipping Fee: $%.2f</li>
		<li>Total: $%.2f</li>
	</ul>
	<p>Thank you for shopping with Beauty Shop ❤️</p>
	`, user.Name, createdOrder.ID.Hex(), createdOrder.DeliveryType, createdOrder.Subtotal, createdOrder.ShippingFee, createdOrder.TotalPrice)

	utils.QueueEmail(user.Email, user.Name, subject, html)
}

// GetOrders handles GET /orders
func GetOrders(c *gin.Context) {
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

	orders, err := orderService.GetOrdersByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetOrderByID handles GET /orders/:id
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

// CancelOrder handles PUT /orders/:id/cancel
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
