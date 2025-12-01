package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// -------------------- PRODUCT CONTROLLER --------------------
type ProductController struct {
	productService services.ProductService
}

var (
	productControllerInstance *ProductController
	once                      sync.Once
)

// Initialize singleton
func InitProductController(productService services.ProductService) {
	once.Do(func() {
		productControllerInstance = &ProductController{productService: productService}
	})
}

// Get singleton instance
func ProductControllerSingleton() *ProductController {
	if productControllerInstance == nil {
		panic("ProductController not initialized. Call InitProductController first!")
	}
	return productControllerInstance
}

// -------------------- CREATE PRODUCT --------------------
func (pc *ProductController) CreateProduct(c *gin.Context) {
	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Category    string  `json:"category"`
		ImageURL    string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	if req.Name == "" || req.Price <= 0 || req.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, price, and stock are required"})
		return
	}

	// Force upload
	fmt.Println("⚙️ Uploading image to Cloudinary:", req.ImageURL)
	cloudImageURL := ""
	if req.ImageURL != "" {
		url, err := utils.UploadRemoteImage(req.ImageURL)
		if err != nil {
			fmt.Println("❌ Cloudinary upload failed:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image: " + err.Error()})
			return
		}
		cloudImageURL = url
		fmt.Println("✅ Cloudinary URL:", cloudImageURL)
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		ImageURL:    cloudImageURL,
	}

	if err := pc.productService.CreateProduct(&product); err != nil {
		fmt.Println("❌ Product save failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("✅ Product created:", product)
	c.JSON(http.StatusCreated, gin.H{"product": product})
}

func (pc *ProductController) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *float64 `json:"price"`
		Stock       *int     `json:"stock"`
		Category    *string  `json:"category"`
		ImageURL    *string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := models.Product{}

	if input.Name != nil {
		update.Name = *input.Name
	}
	if input.Description != nil {
		update.Description = *input.Description
	}
	if input.Price != nil {
		update.Price = *input.Price
	}
	if input.Stock != nil {
		update.Stock = *input.Stock
	}
	if input.Category != nil {
		update.Category = *input.Category
	}
	if input.ImageURL != nil {
		update.ImageURL = *input.ImageURL
	}

	if err := pc.productService.UpdateProduct(id, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (pc *ProductController) GetAllProducts(c *gin.Context) {
	products, err := pc.productService.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (pc *ProductController) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}
