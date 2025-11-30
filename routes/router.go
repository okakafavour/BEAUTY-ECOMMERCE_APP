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
	// AUTH ROUTES (Signup & Login)
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

	// Public routes
	r.GET("/products", productController.GetAllProducts)
	r.GET("/products/:id", productController.GetProductByID)

	// Admin protected routes
	adminRoutes := r.Group("/products")
	adminRoutes.Use(middlewares.JWTMiddleware(), middlewares.AdminMiddleware())
	{
		adminRoutes.POST("", productController.CreateProduct)
		adminRoutes.PUT("/:id", productController.UpdateProduct)
		adminRoutes.DELETE("/:id", productController.DeleteProduct)
	}

	// ----------------------------------------
	// CART ROUTES
	// ----------------------------------------
	cartRepo := repositories.NewCartRepository(db)
	cartService := servicesimpl.NewCartService(cartRepo)
	controllers.InitCartController(cartService)

	cartRoutes := r.Group("/cart")
	cartRoutes.Use(middlewares.JWTMiddleware())
	{
		cartRoutes.POST("", controllers.CreateCart)
		cartRoutes.GET("", controllers.GetCart)
		cartRoutes.PUT("/:id", controllers.UpdateCartItem)
		cartRoutes.DELETE("/:id", controllers.DeleteCartItem)
		cartRoutes.DELETE("", controllers.ClearCart)
	}

	// ----------------------------------------
	// ORDER & PAYMENT ROUTES
	// ----------------------------------------
	orderRepo := repositories.NewOrderRepository(db)
	orderService := servicesimpl.NewOrderService(orderRepo, productRepo)

	// Initialize controllers with orderService
	controllers.InitOrderController(orderService)
	controllers.InitPaymentController(orderService) // MUST be called AFTER orderService exists

	orderRoutes := r.Group("/orders")
	orderRoutes.Use(middlewares.JWTMiddleware())
	{
		orderRoutes.POST("", controllers.CreateOrder)
		orderRoutes.GET("", controllers.GetOrders)
		orderRoutes.GET("/:id", controllers.GetOrderByID)
		orderRoutes.PUT("/:id/cancel", controllers.CancelOrder)
		orderRoutes.POST("/:id/pay", controllers.InitializePayment) // Payment initialization
	}

	// Payment success callback
	r.GET("/payment/success", controllers.PaymentSuccess)

	// ----------------------------------------
	// DEBUG ROUTE (Optional, confirm payment service is initialized)
	// ----------------------------------------
	r.GET("/debug/payment_service", func(c *gin.Context) {
		if controllers.PaymentOrderService == nil {
			c.JSON(500, gin.H{"status": "nil"})
		} else {
			c.JSON(200, gin.H{"status": "initialized"})
		}
	})
}
