package routes

import (
	"beauty-ecommerce-backend/controllers"
	"beauty-ecommerce-backend/middlewares"
	"beauty-ecommerce-backend/repositories"
	servicesimpl "beauty-ecommerce-backend/services_impl"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetUpRoutes(router *gin.Engine, db *mongo.Database) {

	// ===== user routes =====
	controllers.InitUserController()
	router.POST("/signup", controllers.Register)
	router.POST("/login", controllers.Login)

	// ===== product routes =====
	productRepo := repositories.NewProductRepository(db)
	productService := servicesimpl.NewProductService(productRepo)

	// Initialize singleton controller
	controllers.InitProductController(productService)
	productController := controllers.ProductControllerSingleton()

	// Product routes
	products := router.Group("/products")
	{
		// Public routes
		products.GET("/", productController.GetAllProducts)
		products.GET("/:id", productController.GetProductByID)

		// Admin-only routes
		products.Use(middlewares.AdminMiddleware())
		{
			products.POST("/", productController.CreateProduct)
			products.PUT("/:id", productController.UpdateProduct)
			products.DELETE("/:id", productController.DeleteProduct)
		}
	}
}
