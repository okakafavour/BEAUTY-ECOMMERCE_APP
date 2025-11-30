package controllers

import (
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishlistController struct {
	service services.WishlistService // interface, not pointer
}

func NewWishlistController(service services.WishlistService) *WishlistController {
	return &WishlistController{service}
}

// GET /wishlist
func (wc *WishlistController) GetWishlist(c *gin.Context) {
	userID, _ := utils.ExtractUserIDAndRole(c)
	wishlist, err := wc.service.GetWishlist(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wishlist": wishlist})
}

// POST /wishlist/add
func (wc *WishlistController) AddToWishlist(c *gin.Context) {
	userID, _ := utils.ExtractUserIDAndRole(c)

	var req struct {
		ProductID string `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pid, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	err = wc.service.AddProduct(userID, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product added to wishlist"})
}

// POST /wishlist/remove
func (wc *WishlistController) RemoveFromWishlist(c *gin.Context) {
	userID, _ := utils.ExtractUserIDAndRole(c)

	var req struct {
		ProductID string `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pid, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	err = wc.service.RemoveProduct(userID, pid) // ✅ Use RemoveProduct
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product removed from wishlist"})
}
