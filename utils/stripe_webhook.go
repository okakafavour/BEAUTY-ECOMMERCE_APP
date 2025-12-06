package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Example order struct (replace with your real DB model)
type Order struct {
	ID              string
	PaymentIntentID string
	Status          string
}

// Mock DB (replace with your real DB queries)
var ordersDB = map[string]*Order{}

func AddTestOrder(order *Order) {
	ordersDB[order.ID] = order
}

func findOrderByPaymentIntentID(pid string) *Order {
	for _, o := range ordersDB {
		if o.PaymentIntentID == pid {
			return o
		}
	}
	return nil
}

// StripeWebhookHandler handles incoming Stripe webhook events
func StripeWebhookHandler(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusServiceUnavailable, "Error reading request body")
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	sigHeader := c.GetHeader("Stripe-Signature")

	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		log.Println("⚠️ Webhook signature verification failed:", err)
		c.String(http.StatusBadRequest, "Signature verification failed")
		return
	}

	log.Printf("🔥 Received event: %s\n", event.Type)

	switch event.Type {

	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent

		// FIX: Use json.Unmarshal instead of stripe.UnmarshalJSON
		err := json.Unmarshal(event.Data.Raw, &pi)
		if err != nil {
			log.Println("❌ Error parsing PaymentIntent:", err)
			c.Status(http.StatusBadRequest)
			return
		}

		log.Println("✅ Payment succeeded:", pi.ID)

		// Example: fetch order using payment intent ID
		order := findOrderByPaymentIntentID(pi.ID)
		if order == nil {
			log.Println("⚠️ Order not found for this payment reference:", pi.ID)
			break
		}

		if order.Status != "paid" {
			order.Status = "paid"
			log.Printf("🎉 Order %s marked as PAID\n", order.ID)
		} else {
			log.Printf("⚠️ Order %s was ALREADY paid\n", order.ID)
		}

	default:
		log.Println("⚠️ Unhandled event:", event.Type)
	}

	c.Status(http.StatusOK)
}
