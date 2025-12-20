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
		// Local test mode: skip signature verification (Postman)
		fmt.Println("⚡ Local test mode: skipping signature verification")
		if err := json.Unmarshal(payload, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
	} else {
		// Verify real Stripe signature
		event, err = webhook.ConstructEventWithOptions(
			payload,
			sigHeader,
			webhookSecret,
			webhook.ConstructEventOptions{
				IgnoreAPIVersionMismatch: true,
			},
		)
		if err != nil {
			fmt.Println("❌ Signature Verification Failed:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
			return
		}
	}

	fmt.Println("🔥 Received event:", event.Type)
	fmt.Println("⚡ Webhook hit!")

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
		} else {
			fmt.Println("🎉 Order marked as PAID")
		}

		go func() {
			userEmail := pi.Metadata["user_email"]
			userName := pi.Metadata["user_name"]
			orderID := pi.Metadata["order_id"]
			if userEmail == "" {
				userEmail = "okakafavour81@gmail.com" // fallback for Postman test
			}
			if err := utils.SendConfirmationEmail(userEmail, userName, orderID); err != nil {
				fmt.Println("❌ Failed to send confirmation email:", err)
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
		} else {
			fmt.Println("Order marked as FAILED")
		}

		go func() {
			userEmail := pi.Metadata["user_email"]
			userName := pi.Metadata["user_name"]
			orderID := pi.Metadata["order_id"]
			if userEmail == "" {
				userEmail = "okakafavour81@gmail.com" // fallback for Postman test
			}
			if err := utils.SendFailedPaymentEmail(userEmail, userName, orderID); err != nil {
				fmt.Println("❌ Failed to send failed payment email:", err)
			}
		}()

	case "charge.refunded":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err == nil {
			fmt.Println("💸 Payment refunded:", charge.ID)
			if charge.PaymentIntent != nil {
				if err := PaymentOrderService.MarkOrderAsRefunded(charge.PaymentIntent.ID); err != nil {
					fmt.Println("❌ Failed to mark order as refunded:", err)
				} else {
					fmt.Println("🎉 Order marked as REFUNDED")
				}
			} else {
				fmt.Println("❌ No PaymentIntent associated with this charge")
			}
		}

	case "charge.dispute.created":
		var dispute stripe.Dispute
		if err := json.Unmarshal(event.Data.Raw, &dispute); err == nil {

			if dispute.Charge == nil {
				fmt.Println("❌ dispute.Charge is nil")
				break
			}

			chargeID := dispute.Charge.ID // THIS is the actual string

			fmt.Println("⚠️ Payment disputed, Charge ID:", chargeID)

			if err := PaymentOrderService.MarkOrderAsDisputed(chargeID); err != nil {
				fmt.Println("❌ Failed to mark order as disputed:", err)
			} else {
				fmt.Println("🎉 Order marked as DISPUTED")
			}
		}

	default:
		fmt.Println("⚠️ Unhandled event:", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
