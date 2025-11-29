package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var cartService services.CartService

func InitCartController(cs services.CartService) {
	cartService = cs
}
func CreateCart(c *gin.Context) {
	var body struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get userID from middleware (STRING, not ObjectID)
	userIDHex, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userIDHex.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id must be a string"})
		return
	}

	// Convert string → ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	// Convert productID to ObjectID
	productID, err := primitive.ObjectIDFromHex(body.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	cartItem := models.CartItem{
		ID:        primitive.NewObjectID(),
		ProductID: productID,
		UserID:    userID,
		Quantity:  body.Quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdCartItem, err := cartService.CreateCartItem(cartItem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cart_item": createdCartItem})
}

func GetCart(c *gin.Context) {
	userIDHex, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userIDHex.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id must be a string"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	cartItems, err := cartService.GetCartByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cart_items": cartItems})
}

func UpdateCartItem(c *gin.Context) {
	cartItemID, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var body struct {
		Quantity int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// STEP 1 — Get existing cart item
	existing, err := cartService.GetCartItemByID(cartItemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	// STEP 2 — Update quantity only
	existing.Quantity = body.Quantity
	existing.UpdatedAt = time.Now()

	updated, err := cartService.UpdateCartItem(*existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cart_item": updated})
}

func DeleteCartItem(c *gin.Context) {
	cartItemID, _ := primitive.ObjectIDFromHex(c.Param("id"))
	err := cartService.DeleteCartItem(cartItemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted"})
}

func ClearCart(c *gin.Context) {
	userIDHex, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userIDHex.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id must be a string"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	cartItems, err := cartService.GetCartByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(cartItems) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Cart is already empty"})
		return
	}

	for _, item := range cartItems {
		if err := cartService.DeleteCartItem(item.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
}
