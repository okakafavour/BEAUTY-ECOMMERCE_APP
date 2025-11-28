package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService services.ProductService
}

func NewProductController(productService services.ProductService) *ProductController {
	return &ProductController{productService: productService}
}

var (
	productControllerInstance *ProductController
	once                      sync.Once
)

func InitProductController(productService services.ProductService) {
	once.Do(func() {
		productControllerInstance = NewProductController(productService)
	})
}

func ProductControllerSingleton() *ProductController {
	if productControllerInstance == nil {
		panic("ProductController not initialized. Call InitProductController first!")
	}
	return productControllerInstance
}

// ------------------ CREATE ------------------
func (pc *ProductController) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pass pointer to service so ID & timestamps are set correctly
	if err := pc.productService.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": product, // now has correct ID
	})
}

// ------------------ GET ALL ------------------
func (pc *ProductController) GetAllProducts(c *gin.Context) {
	products, err := pc.productService.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var resp []gin.H
	for _, p := range products {
		resp = append(resp, gin.H{
			"id":          p.ID.Hex(),
			"name":        p.Name,
			"description": p.Description,
			"price":       p.Price,
			"category":    p.Category,
			"image_url":   p.ImageURL,
			"stock":       p.Stock,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"products": resp})
}

// ------------------ GET BY ID ------------------
func (pc *ProductController) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{
		"id":          product.ID.Hex(),
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"category":    product.Category,
		"image_url":   product.ImageURL,
		"stock":       product.Stock,
		"created_at":  product.CreatedAt,
		"updated_at":  product.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"product": resp})
}

// ======= update product =======
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := pc.productService.UpdateProduct(id, product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch the updated product to return it
	updatedProduct, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": updatedProduct,
	})
}

// ------------------ DELETE ------------------
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := pc.productService.DeleteProduct(id)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Product deleted successfully",
		"deleted_id": id,
	})
}
