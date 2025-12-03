package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"beauty-ecommerce-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------
// Connect to test MongoDB
func ConnectTestDB() (*mongo.Database, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	db := client.Database("ecommerce_test")
	return db, ctx, cancel
}

// ---------------------------
// SetupRouter with fresh controller
func SetupRouter(db *mongo.Database) *gin.Engine {
	r := gin.Default()

	// // Create product service and controller directly
	// productService := servicesimpl.NewProductService(repositories.NewProductRepository(db))
	// // pc := NewProductController(productService)

	// // Register routes
	// r.POST("/products", pc.CreateProduct)
	// r.GET("/products", pc.GetAllProducts)
	// r.GET("/products/:id", pc.GetProductByID)
	// r.PUT("/products/:id", pc.UpdateProduct)
	// r.DELETE("/products/:id", pc.DeleteProduct)

	return r
}

// ---------------------------
// Test: Create product
func TestCreateProduct(t *testing.T) {
	db, ctx, cancel := ConnectTestDB()
	defer cancel()
	defer db.Drop(ctx)

	r := SetupRouter(db)

	product := models.Product{
		Name:        "Test Product",
		Description: "This is a test product",
		Price:       10.5,
		Stock:       5,
		Category:    "Test Category",
		ImageURL:    "http://example.com/image.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jsonValue, _ := json.Marshal(product)
	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, product.Name, resp["product"].(map[string]interface{})["name"])
}

// ---------------------------
// Test: Delete product
func TestDeleteProduct(t *testing.T) {
	db, ctx, cancel := ConnectTestDB()
	defer cancel()
	defer db.Drop(ctx)

	r := SetupRouter(db)

	// 1. Create product
	product := models.Product{
		Name:        "Test Product",
		Description: "Test",
		Price:       10,
		Stock:       2,
		Category:    "Test",
		ImageURL:    "http://img.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jsonValue, _ := json.Marshal(product)
	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Role", "admin")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Extract REAL ID returned by API
	var created map[string]interface{}
	json.Unmarshal([]byte(w.Body.String()), &created)

	id := created["product"].(map[string]interface{})["id"].(string)

	// 3. Delete with REAL ID
	req2, _ := http.NewRequest("DELETE", "/products/"+id, nil)
	req2.Header.Set("Role", "admin")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestUpdateProduct(t *testing.T) {
	db, ctx, cancel := ConnectTestDB()
	defer cancel()
	defer db.Drop(ctx)

	r := SetupRouter(db)

	// 1. Create product first
	product := models.Product{
		Name:        "Original Product",
		Description: "Original description",
		Price:       10.0,
		Stock:       5,
		Category:    "Original Category",
		ImageURL:    "http://example.com/original.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jsonValue, _ := json.Marshal(product)
	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Extract the created product ID
	var created map[string]interface{}
	json.Unmarshal([]byte(w.Body.String()), &created)
	id := created["product"].(map[string]interface{})["id"].(string)

	// 3. Prepare update payload
	update := models.Product{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       20.0,
		Stock:       10,
		Category:    "Updated Category",
		ImageURL:    "http://example.com/updated.jpg",
	}

	jsonUpdate, _ := json.Marshal(update)
	req2, _ := http.NewRequest("PUT", "/products/"+id, bytes.NewBuffer(jsonUpdate))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 4. Fetch the product to confirm changes
	req3, _ := http.NewRequest("GET", "/products/"+id, nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	var updated map[string]interface{}
	json.Unmarshal([]byte(w3.Body.String()), &updated)
	prod := updated["product"].(map[string]interface{})
	assert.Equal(t, "Updated Product", prod["name"])
	assert.Equal(t, "Updated description", prod["description"])
	assert.Equal(t, float64(20.0), prod["price"])
}
