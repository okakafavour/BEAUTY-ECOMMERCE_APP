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

	// Get userID from middleware
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDRaw.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Convert productID from string to ObjectID
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
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDRaw.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
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
	c.ShouldBindJSON(&body)

	cartItem := models.CartItem{ID: cartItemID, Quantity: body.Quantity}
	updated, err := cartService.UpdateCartItem(cartItem)
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
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDRaw.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
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
