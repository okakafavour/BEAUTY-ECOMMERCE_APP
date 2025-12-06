package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var cartService services.CartService

func InitCartController(cs services.CartService) {
	cartService = cs
}

// -------------------- Helper --------------------
func getUserIDFromContext(c *gin.Context) (primitive.ObjectID, error) {
	userClaims, exists := c.Get("user")
	if !exists {
		return primitive.NilObjectID, fmt.Errorf("Unauthorized")
	}

	claimsMap, ok := userClaims.(jwt.MapClaims)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("Cannot read token claims")
	}

	userIDStr, ok := claimsMap["user_id"].(string)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("user_id must be a string")
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("Invalid user_id format")
	}

	return userID, nil
}

// -------------------- CREATE CART ITEM --------------------
func CreateCart(c *gin.Context) {
	var body struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

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

// -------------------- GET CART ITEMS --------------------
func GetCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cartItems, err := cartService.GetCartByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cart_items": cartItems})
}

// -------------------- UPDATE CART ITEM --------------------
func UpdateCartItem(c *gin.Context) {
	cartItemID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
		return
	}

	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	existing, err := cartService.GetCartItemByID(cartItemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	existing.Quantity = body.Quantity
	existing.UpdatedAt = time.Now()

	updated, err := cartService.UpdateCartItem(*existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cart_item": updated})
}

// -------------------- DELETE CART ITEM --------------------
func DeleteCartItem(c *gin.Context) {
	cartItemID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
		return
	}

	if err := cartService.DeleteCartItem(cartItemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted"})
}

// -------------------- CLEAR CART --------------------
func ClearCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cartItems, err := cartService.GetCartByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
