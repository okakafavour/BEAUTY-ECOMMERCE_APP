package controllers

import (
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewController struct {
	service *services.ReviewService
}

func NewReviewController(service *services.ReviewService) *ReviewController {
	return &ReviewController{service}
}

// CreateReview creates a review with the correct user_id
func (rc *ReviewController) CreateReview(c *gin.Context) {
	userID, _ := utils.ExtractUserIDAndRole(c) // ✅ get real user ID

	var req struct {
		ProductID string `json:"product_id"`
		Rating    int    `json:"rating"`
		Body      string `json:"body"`
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

	err = rc.service.CreateReview(userID, pid, req.Rating, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "review created"})
}

// Get reviews for a product
func (rc *ReviewController) GetProductReviews(c *gin.Context) {
	productID := c.Param("productId")
	pid, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	reviews, err := rc.service.GetProductReviews(pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

// ---------------- Update Review ----------------
func (rc *ReviewController) UpdateReview(c *gin.Context) {
	userID, role := utils.ExtractUserIDAndRole(c)
	isAdmin := role == "ADMIN"

	reviewID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		Rating int    `json:"rating"`
		Body   string `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = rc.service.UpdateReview(reviewID, userID, isAdmin, req.Rating, req.Body)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review updated"})
}

// ---------------- Delete Review ----------------
func (rc *ReviewController) DeleteReview(c *gin.Context) {
	userID, role := utils.ExtractUserIDAndRole(c)
	isAdmin := role == "ADMIN"

	reviewID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	err = rc.service.DeleteReview(reviewID, userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review deleted"})
}
