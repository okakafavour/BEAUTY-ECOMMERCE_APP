package controllers

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService services.ProductService
}

var (
	productControllerInstance *ProductController
	once                      sync.Once
)

func InitProductController(productService services.ProductService) {
	once.Do(func() {
		productControllerInstance = &ProductController{productService: productService}
	})
}

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
	_ = c.ShouldBindJSON(&req)

	// -------------------- Read form-data values (overwrite JSON) --------------------
	name := req.Name
	description := req.Description
	price := req.Price
	stock := req.Stock
	category := req.Category

	if formName := c.PostForm("name"); formName != "" {
		name = formName
	}
	if formDescription := c.PostForm("description"); formDescription != "" {
		description = formDescription
	}
	if formPrice := c.PostForm("price"); formPrice != "" {
		fmt.Sscanf(formPrice, "%f", &price)
	}
	if formStock := c.PostForm("stock"); formStock != "" {
		fmt.Sscanf(formStock, "%d", &stock)
	}
	if formCategory := c.PostForm("category"); formCategory != "" {
		category = formCategory
	}

	if name == "" || price <= 0 || stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, price, and stock are required"})
		return
	}

	var imageURL, imageID string

	file, err := c.FormFile("image")
	if err == nil {
		url, publicID, err := utils.UploadToCloudinaryWithID(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image: " + err.Error()})
			return
		}
		imageURL = url
		imageID = publicID
	} else if req.ImageURL != "" {
		url, publicID, err := utils.UploadRemoteImageWithID(req.ImageURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload remote image: " + err.Error()})
			return
		}
		imageURL = url
		imageID = publicID
	}

	if imageURL != "" && imageID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve Cloudinary public ID"})
		return
	}

	fmt.Println("DEBUG: ImageURL:", imageURL, "ImageID:", imageID)

	fmt.Println("DEBUG BEFORE CREATE:")
	fmt.Println("Name:", name)
	fmt.Println("ImageURL:", imageURL)
	fmt.Println("ImageID:", imageID)

	product := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Category:    category,
		ImageURL:    imageURL,
		ImageID:     imageID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	fmt.Println("✅ Saving Product:", product.Name, "ImageID:", imageID)

	if err := pc.productService.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	_ = c.ShouldBindJSON(&input)

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

	file, err := c.FormFile("image")
	if err == nil {
		url, _, err := utils.UploadToCloudinaryWithID(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image: " + err.Error()})
			return
		}
		update.ImageURL = url
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

// -------------------- DELETE PRODUCT --------------------
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	// ------------------- DELETE IMAGE -------------------
	if product.ImageID != "" {
		// Preferred: delete by public ID
		if err := utils.DeleteImageFromCloudinary(product.ImageID); err != nil {
			fmt.Println("❌ Failed to delete image by ImageID:", err)
		} else {
			fmt.Println("✅ Deleted image from Cloudinary by ImageID:", product.ImageID)
		}
	} else if product.ImageURL != "" {
		// Fallback: extract public ID from URL and delete
		if err := utils.DeleteImageFromCloudinaryByURL(product.ImageURL); err != nil {
			fmt.Println("❌ Failed to delete image by ImageURL:", err)
		} else {
			fmt.Println("✅ Deleted image from Cloudinary by URL:", product.ImageURL)
		}
	}

	// ------------------- DELETE PRODUCT -------------------
	if err := pc.productService.DeleteProduct(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "product deleted successfully",
		"productId": product.ID.Hex(),
		"name":      product.Name,
	})
}
