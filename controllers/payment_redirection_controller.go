package controllers

import (
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Global OrderService
var PaymentOrderService services.OrderService

// Initialize the controller with the OrderService and Stripe key
func InitPaymentController(s services.OrderService) {
	PaymentOrderService = s
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// POST /payment/initialize/:id
func InitializePayment(c *gin.Context) {
	if PaymentOrderService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment service not initialized"})
		return
	}

	orderIDHex := c.Param("id")
	orderID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	// Fetch order
	order, err := PaymentOrderService.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order not found"})
		return
	}

	if order.Total <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order total"})
		return
	}

	// Create Stripe PaymentIntent using utils function
	pi, err := utils.CreateStripePaymentIntent(order.Total, "ngn", orderIDHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save PaymentIntent ID in order
	if err := PaymentOrderService.SaveOrderReference(orderIDHex, pi.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save payment reference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "payment initialized",
		"order_id":      orderIDHex,
		"amount":        order.Total,
		"client_secret": pi.ClientSecret,
		"payment_id":    pi.ID,
	})
}

// POST /payment/webhook
func StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook signature verification failed: " + err.Error()})
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment intent object"})
			return
		}

		orderIDHex := pi.Metadata["order_id"]

		// Mark order as paid
		if err := PaymentOrderService.MarkOrderAsPaid(orderIDHex); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Fetch order details
		orderIDObj, err := primitive.ObjectIDFromHex(orderIDHex)
		if err != nil {
			fmt.Println("invalid order ID:", err)
			break
		}

		order, err := PaymentOrderService.GetOrderByID(orderIDObj)
		if err != nil {
			fmt.Println("failed to fetch order:", err)
			break
		}

		// Fetch user details (to get user name)
		user, err := userService.GetUserByID(order.UserID)
		if err != nil {
			fmt.Println("failed to fetch user:", err)
			break
		}

		// Send payment success email
		orderIDStr := order.ID.Hex()
		subject, html := utils.PaymentSuccessEmail(user.Name, orderIDStr)
		err = utils.SendEmail(user.Email, subject, "", html)
		if err != nil {
			fmt.Println("failed to send payment success email:", err)
		}

	default:
		// ignore other events
		c.JSON(http.StatusOK, gin.H{"message": "event ignored"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
