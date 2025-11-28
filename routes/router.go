package routes

import (
	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/controllers"
	"beauty-ecommerce-backend/middlewares"
	"beauty-ecommerce-backend/repositories"
	servicesimpl "beauty-ecommerce-backend/services_impl"

	"github.com/gin-gonic/gin"
)

func SetUpRoutes(r *gin.Engine) {

	// ----------------------------------------
	// AUTH ROUTES  (Signup & Login)
	// ----------------------------------------
	controllers.InitUserController()

	r.POST("/signup", controllers.Register)
	r.POST("/login", controllers.Login)

	// ----------------------------------------
	// PRODUCT ROUTES
	// ----------------------------------------
	db := config.DB
	productRepo := repositories.NewProductRepository(db)
	productService := servicesimpl.NewProductService(productRepo)
	productController := controllers.NewProductController(productService)

	// Public
	r.GET("/products", productController.GetAllProducts)
	r.GET("/products/:id", productController.GetProductByID)

	// Admin Protected
	adminRoutes := r.Group("/products")
	adminRoutes.Use(middlewares.JWTMiddleware(), middlewares.AdminMiddleware())
	{
		adminRoutes.POST("", productController.CreateProduct)
		adminRoutes.PUT("/:id", productController.UpdateProduct)
		adminRoutes.DELETE("/:id", productController.DeleteProduct)
	}
}
