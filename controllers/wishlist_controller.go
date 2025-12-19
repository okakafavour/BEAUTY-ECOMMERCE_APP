package controllers

import (
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishlistController struct {
	service services.WishlistService
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

// GET /wishlist?page=1&limit=10
func (wc *WishlistController) GetWishlistPaginated(c *gin.Context) {
	userID, _ := utils.ExtractUserIDAndRole(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, total, err := wc.service.GetWishlistPaginated(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}
