package controllers

import (
	"beauty-ecommerce-backend/config"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// -------------------------------
// TestMain runs before all tests
// -------------------------------
func TestMain(m *testing.M) {
	// Load .env
	if err := godotenv.Load("../.env"); err != nil {
		panic("❌ Error loading .env file")
	}

	// Connect to test DB
	config.ConnectTestDB()

	// Initialize controller
	// InitUserController()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// -------------------------------
// Helpers
// -------------------------------
func cleanUsersCollection() {
	if config.DB == nil {
		panic("Database not initialized! Did you run TestMain?")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := config.DB.Collection("users").Drop(ctx); err != nil {
		panic(err)
	}
}

func setUpUserRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/signup", Register)
	r.POST("/login", Login)

	return r
}

// -------------------------------
// Tests
// -------------------------------
func TestToRegisterUser(t *testing.T) {
	cleanUsersCollection()
	router := setUpUserRouter()

	body := []byte(`{
		"name": "Okaka Favour",
		"email": "okaka@example.com",
		"password": "12345"
	}`)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, "User created successfully", resp["message"])
}

func TestToRegisterWithoutEmail(t *testing.T) {
	cleanUsersCollection()
	router := setUpUserRouter()

	body := []byte(`{"name": "Sam", "password": "0000"}`)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Email is required", resp["error"])
}

func TestToRegisterDuplicateEmail(t *testing.T) {
	cleanUsersCollection()
	router := setUpUserRouter()

	body := []byte(`{
		"name": "Favour",
		"email": "dup@example.com",
		"password": "123"
	}`)

	// First signup
	req1, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Duplicate signup
	req2, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.Equal(t, "email already registered", resp["error"])
}

func TestToLoginUser(t *testing.T) {
	cleanUsersCollection()
	router := setUpUserRouter()

	signupBody := []byte(`{
		"name": "Okaka Favour",
		"email": "okaka@example.com",
		"password": "12345"
	}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(signupBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	loginBody := []byte(`{
		"email": "okaka@example.com",
		"password": "12345"
	}`)
	req2, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(loginBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.Equal(t, "Login successful", resp["message"])
	assert.NotEmpty(t, resp["token"])
}

func TestToLoginInvalidPassword(t *testing.T) {
	cleanUsersCollection()
	router := setUpUserRouter()

	body := []byte(`{
		"name": "Okaka Favour",
		"email": "okaka@example.com",
		"password": "12345"
	}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	loginBody := []byte(`{
		"email": "okaka@example.com",
		"password": "wrongpass"
	}`)
	req2, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(loginBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.Equal(t, "invalid email or password", resp["error"])
}
