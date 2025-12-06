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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	PaymentOrderService services.OrderService
	PaymentUserService  services.UserService
)

// InitPaymentController initializes services and Stripe key
func InitPaymentController(orderService services.OrderService, userService services.UserService) {
	PaymentOrderService = orderService
	PaymentUserService = userService
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// POST /orders/:id/pay
func InitializePayment(c *gin.Context) {
	orderIDHex := c.Param("id")
	orderID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := PaymentOrderService.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order not found"})
		return
	}

	if order.Total <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order total invalid"})
		return
	}

	user, err := PaymentUserService.GetUserByID(order.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	// Create Stripe PaymentIntent with metadata
	pi, err := utils.CreateStripePaymentIntentWithMetadata(order.Total, "ngn", orderIDHex, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save PaymentIntent ID to order
	err = PaymentOrderService.SaveOrderReference(orderIDHex, pi.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save payment reference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "payment initialized",
		"order_id":      orderIDHex,
		"amount":        order.Total,
		"payment_id":    pi.ID,
		"client_secret": pi.ClientSecret,
	})
}

// POST /payment/webhook
// Signature verification removed for Postman testing
func StripeWebhook(c *gin.Context) {

	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Parse event manually (NO SIGNATURE CHECK)
	var event stripe.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Println("❌ Failed to parse webhook:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook structure"})
		return
	}

	fmt.Println("🔥 Webhook Event:", event.Type)

	switch event.Type {

	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			fmt.Println("❌ Error parsing PaymentIntent:", err)
			break
		}

		fmt.Println("✅ Payment succeeded:", pi.ID)

		if err := PaymentOrderService.MarkOrderAsPaid(pi.ID); err != nil {
			fmt.Println("❌ Failed to mark order as paid:", err)
		} else {
			fmt.Println("🎉 Order marked as PAID")
		}

		// send the email
		utils.SendConfirmationEmail(
			pi.Metadata["user_email"],
			pi.Metadata["user_name"],
			pi.Metadata["order_id"],
		)

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			fmt.Println("❌ Error parsing failed PaymentIntent:", err)
			break
		}

		fmt.Println("❌ Payment FAILED:", pi.ID)

		if err := PaymentOrderService.MarkOrderAsFailed(pi.ID); err != nil {
			fmt.Println("❌ Failed to update order:", err)
		}

		utils.SendFailedPaymentEmail(
			pi.Metadata["user_email"],
			pi.Metadata["user_name"],
			pi.Metadata["order_id"],
		)

	default:
		fmt.Println("⚠️ Unhandled event:", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
