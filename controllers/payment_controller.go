package controllers

import (
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
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

	if order.TotalPrice <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order total invalid"})
		return
	}

	user, err := PaymentUserService.GetUserByID(order.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	pi, err := utils.CreateStripePaymentIntentWithMetadata(order.TotalPrice, orderIDHex, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = PaymentOrderService.SaveOrderReference(orderIDHex, pi.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save payment reference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "payment initialized",
		"order_id":      orderIDHex,
		"amount":        order.TotalPrice,
		"payment_id":    pi.ID,
		"client_secret": pi.ClientSecret,
	})
}

// POST /payment/webhook
func StripeWebhook(c *gin.Context) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook secret not configured"})
		return
	}

	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	var event stripe.Event

	if sigHeader == "" {
		// Local test mode
		fmt.Println("⚡ Local test mode: skipping signature verification")
		if err := json.Unmarshal(payload, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
	} else {
		event, err = webhook.ConstructEventWithOptions(
			payload, sigHeader, webhookSecret,
			webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
		)
		if err != nil {
			fmt.Println("❌ Signature Verification Failed:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
			return
		}
	}

	fmt.Println("🔥 Received event:", event.Type)
	fmt.Println("⚡ Webhook hit!")

	adminEmail := os.Getenv("ADMIN_EMAIL")

	switch event.Type {

	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			fmt.Println("❌ Failed to parse PaymentIntent:", err)
			break
		}

		fmt.Println("✅ Payment succeeded:", pi.ID)
		if err := PaymentOrderService.MarkOrderAsPaid(pi.ID); err != nil {
			fmt.Println("❌ Failed to mark order as paid:", err)
		}

		go func() {
			// Read metadata
			userEmail := pi.Metadata["user_email"]
			userName := pi.Metadata["user_name"]
			orderID := pi.Metadata["order_id"]
			deliveryType := pi.Metadata["delivery_type"]
			subtotalStr := pi.Metadata["subtotal"]
			shippingStr := pi.Metadata["shipping_fee"]
			totalStr := pi.Metadata["total_price"]

			if userEmail == "" {
				userEmail = "okakafavour81@gmail.com" // fallback
			}

			// Convert strings to float64
			subtotal, _ := strconv.ParseFloat(subtotalStr, 64)
			shipping, _ := strconv.ParseFloat(shippingStr, 64)
			total, _ := strconv.ParseFloat(totalStr, 64)

			// Customer email
			if err := utils.SendConfirmationEmail(userEmail, userName, orderID, deliveryType, subtotal, shipping, total); err != nil {
				fmt.Println("❌ Failed to send customer email:", err)
			}

			// Admin email
			if adminEmail != "" {
				subject := fmt.Sprintf("New Order Paid - %s", orderID)
				html := fmt.Sprintf(`
					<p>Order <b>%s</b> paid by <b>%s</b> (%s).</p>
					<p>Delivery: %s</p>
					<p>Subtotal: £%.2f | Shipping: £%.2f | Total: £%.2f</p>`,
					orderID, userName, userEmail, deliveryType, subtotal, shipping, total)
				if err := utils.SendEmail(adminEmail, subject, "", html); err != nil {
					fmt.Println("❌ Failed to send admin notification:", err)
				}
			}
		}()

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			fmt.Println("❌ Failed to parse failed PaymentIntent:", err)
			break
		}

		fmt.Println("❌ Payment FAILED:", pi.ID)
		if err := PaymentOrderService.MarkOrderAsFailed(pi.ID); err != nil {
			fmt.Println("❌ Failed to mark order as FAILED:", err)
		}

		go func() {
			userEmail := pi.Metadata["user_email"]
			userName := pi.Metadata["user_name"]
			orderID := pi.Metadata["order_id"]
			if userEmail == "" {
				userEmail = "okakafavour81@gmail.com"
			}

			if err := utils.SendFailedPaymentEmail(userEmail, userName, orderID); err != nil {
				fmt.Println("❌ Failed to send customer failed email:", err)
			}

			if adminEmail != "" {
				subject := fmt.Sprintf("Payment FAILED - %s", orderID)
				html := fmt.Sprintf("<p>Payment for order <b>%s</b> by <b>%s</b> (%s) FAILED.</p>", orderID, userName, userEmail)
				if err := utils.SendEmail(adminEmail, subject, "", html); err != nil {
					fmt.Println("❌ Failed to send admin failed notification:", err)
				}
			}
		}()

	// Handle refunds and disputes the same way as before
	case "charge.refunded":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err == nil && charge.PaymentIntent != nil {
			fmt.Println("💸 Payment refunded:", charge.ID)
			if err := PaymentOrderService.MarkOrderAsRefunded(charge.PaymentIntent.ID); err != nil {
				fmt.Println("❌ Failed to mark order as refunded:", err)
			}
		}

	case "charge.dispute.created":
		var dispute stripe.Dispute
		if err := json.Unmarshal(event.Data.Raw, &dispute); err == nil && dispute.Charge != nil {
			fmt.Println("⚠️ Payment disputed, Charge ID:", dispute.Charge.ID)
			if err := PaymentOrderService.MarkOrderAsDisputed(dispute.Charge.ID); err != nil {
				fmt.Println("❌ Failed to mark order as disputed:", err)
			}
		}

	default:
		fmt.Println("⚠️ Unhandled event:", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
