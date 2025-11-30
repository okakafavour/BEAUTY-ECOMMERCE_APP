package controllers

import (
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Global OrderService
var PaymentOrderService services.OrderService

// InitPaymentController sets the global PaymentOrderService
func InitPaymentController(s services.OrderService) {
	PaymentOrderService = s
}

func InitializePayment(c *gin.Context) {
	if PaymentOrderService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment service not initialized"})
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	// Order ID
	orderIDHex := c.Param("id")
	orderID, err := primitive.ObjectIDFromHex(orderIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	// Fetch order to get amount
	order, err := PaymentOrderService.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order not found"})
		return
	}

	if order.Total <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order total"})
		return
	}

	// Generate reference
	reference := primitive.NewObjectID().Hex()

	// Save ref in DB
	if err := PaymentOrderService.SaveOrderReference(orderIDHex, reference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save order reference"})
		return
	}

	// Initialize Paystack WITH THE REAL ORDER AMOUNT
	payRes, err := utils.PaystackInitialize(req.Email, float64(order.Total), reference)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "payment initialized",
		"order_id":    orderIDHex,
		"amount":      order.Total,
		"reference":   reference,
		"payment_url": payRes.Data.AuthorizationURL,
		"access_code": payRes.Data.AccessCode,
	})
}

// GET /payment/success?reference=
func PaymentSuccess(c *gin.Context) {
	if PaymentOrderService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment service not initialized"})
		return
	}

	ref := c.Query("reference")
	if ref == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing reference"})
		return
	}

	// Verify Paystack payment
	_, err := utils.VerifyPaystackPayment(ref)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment verification failed"})
		return
	}

	// Mark order as paid
	if err := PaymentOrderService.MarkOrderAsPaid(ref); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "payment successful",
		"reference": ref,
	})
}
