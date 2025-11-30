package controllers

import (
	"beauty-ecommerce-backend/services"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var paystackOrderService services.OrderService

func InitPaystackWebhookController(s services.OrderService) {
	paystackOrderService = s
}

func PaystackWebhook(c *gin.Context) {
	secret := os.Getenv("PAYSTACK_SECRET_KEY")

	signature := c.GetHeader("X-Paystack-Signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid body"})
		return
	}

	hash := hmac.New(sha512.New, []byte(secret))
	hash.Write(bodyBytes)
	expectedSignature := fmt.Sprintf("%x", hash.Sum(nil))

	if expectedSignature != signature {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	var event map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &event); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}

	if event["event"] != "charge.success" {
		c.JSON(200, gin.H{"message": "Event ignored"})
		return
	}

	data, ok := event["data"].(map[string]interface{})
	if !ok {
		c.JSON(400, gin.H{"error": "Invalid event data"})
		return
	}

	reference, ok := data["reference"].(string)
	if !ok {
		c.JSON(400, gin.H{"error": "Reference not found"})
		return
	}

	err = paystackOrderService.MarkOrderAsPaid(reference)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Payment verified"})
}
