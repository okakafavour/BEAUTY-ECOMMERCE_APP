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
		Page      int    `json:"page"`  // optional
		Limit     int    `json:"limit"` // optional
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

	if err := wc.service.AddProduct(userID, pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return updated wishlist paginated
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 10
	}

	products, total, err := wc.service.GetWishlistPaginated(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "product added to wishlist",
		"products": products,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// POST /wishlist/remove
func (wc *WishlistController) RemoveFromWishlist(c *gin.Context) {
	userID, _ := utils.ExtractUserIDAndRole(c)

	var req struct {
		ProductID string `json:"product_id"`
		Page      int    `json:"page"`  // optional
		Limit     int    `json:"limit"` // optional
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

	if err := wc.service.RemoveProduct(userID, pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return updated wishlist paginated
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 10
	}

	products, total, err := wc.service.GetWishlistPaginated(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "product removed from wishlist",
		"products": products,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
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
