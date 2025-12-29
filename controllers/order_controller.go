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
func CreateOrder(c *gin.Context) {
	var order models.Order

	// Bind JSON
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate DeliveryType
	if order.DeliveryType != "standard" && order.DeliveryType != "express" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery type"})
		return
	}

	// Set shipping fee
	switch order.DeliveryType {
	case "standard":
		order.ShippingFee = 3.99
	case "express":
		order.ShippingFee = 4.99
	}

	// Get user ID from context
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

	// Calculate subtotal from items
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

	// Create order in DB
	createdOrder, err := orderService.CreateOrder(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdOrder)

	// Fetch user info for email
	user, err := userService.GetUserByID(userID)
	if err != nil {
		log.Println("⚠️ Failed to fetch user for order email:", err)
		return
	}

	// Send order confirmation email asynchronously
	go func(order models.Order, userEmail, userName string) {
		subject := fmt.Sprintf("Order Confirmation - %s", order.ID.Hex())
		html := fmt.Sprintf(`
			<h2>Hello %s 👋</h2>
			<p>Your order <b>%s</b> has been received successfully!</p>
			<ul>
				<li>Delivery Type: %s</li>
				<li>Subtotal: $%.2f</li>
				<li>Shipping Fee: $%.2f</li>
				<li>Total: $%.2f</li>
			</ul>
			<p>Thank you for shopping with us!</p>
		`, userName, order.ID.Hex(), order.DeliveryType, order.Subtotal, order.ShippingFee, order.TotalPrice)

		if err := utils.SendMailSenderEmail(userEmail, userName, subject, html); err != nil {
			log.Println("⚠️ Order confirmation email failed:", err)
		} else {
			log.Println("✅ Order confirmation email sent to", userEmail)
		}
	}(createdOrder, user.Email, user.Name)
}

// GET /orders
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
