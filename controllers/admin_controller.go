package controllers

import (
	"net/http"
	"strconv"

	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// -----------------------------
// AdminController Struct
// -----------------------------
type AdminController struct {
	ProductService services.ProductService
	OrderService   services.OrderService
	UserService    services.UserService
}

func NewAdminController(ps services.ProductService, os services.OrderService, us services.UserService) *AdminController {
	return &AdminController{
		ProductService: ps,
		OrderService:   os,
		UserService:    us,
	}
}

// PRODUCT METHODS
func (ac *AdminController) CreateProduct(c *gin.Context) {
	// Parse form-data fields
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	category := c.PostForm("category")

	if name == "" || priceStr == "" || stockStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, price, and stock are required"})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock"})
		return
	}

	imageURL := ""
	imageID := ""

	file, err := c.FormFile("image")
	if err == nil && file != nil {
		uploadedURL, uploadedID, err := utils.UploadToCloudinaryWithID(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image: " + err.Error()})
			return
		}
		imageURL = uploadedURL
		imageID = uploadedID
	}

	// Create product object
	product := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Category:    category,
		ImageURL:    imageURL,
		ImageID:     imageID,
	}

	if err := ac.ProductService.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product": product})
}

func (ac *AdminController) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ac.ProductService.UpdateProduct(id, product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated"})
}

func (ac *AdminController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := ac.ProductService.DeleteProduct(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

//===== ORDER METHODS =====//

func (ac *AdminController) ListOrders(c *gin.Context) {
	orders, err := ac.OrderService.GetAllOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (ac *AdminController) UpdateOrderStatus(c *gin.Context) {
	// Get order ID from URL
	idStr := c.Param("id")
	orderID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	// Bind JSON payload
	var payload struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service with primitive.ObjectID
	if err := ac.OrderService.UpdateOrderStatus(orderID, payload.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}

//////////////////////////////
// USER METHODS
//////////////////////////////

func (ac *AdminController) ListUsers(c *gin.Context) {
	users, err := ac.UserService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (ac *AdminController) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	// Bind JSON payload
	var payload struct {
		Name  string `json:"name,omitempty"`
		Email string `json:"email,omitempty"`
		Role  string `json:"role,omitempty"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build models.User object for update
	update := models.User{
		Name:  payload.Name,
		Email: payload.Email,
		Role:  payload.Role,
	}

	// Call service
	if err := ac.UserService.UpdateUser(id, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (ac *AdminController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := ac.UserService.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

//////////////////////////////
// ANALYTICS (Optional)
//////////////////////////////

func (ac *AdminController) SalesAnalytics(c *gin.Context) {
	data, err := ac.OrderService.GetSalesAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": data})
}
