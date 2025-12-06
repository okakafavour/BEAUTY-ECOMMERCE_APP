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
	db := config.DB

	// --------------------------
	// REPOSITORIES
	// --------------------------
	productRepo := repositories.NewProductRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	userRepo := repositories.NewUserRepository(db)
	cartRepo := repositories.NewCartRepository(db)

	// --------------------------
	// SERVICES
	// --------------------------
	productService := servicesimpl.NewProductService(productRepo)
	userService := servicesimpl.NewUserService(userRepo) // pass userRepo here
	orderService := servicesimpl.NewOrderService(orderRepo, productRepo, userRepo)
	cartService := servicesimpl.NewCartService(cartRepo)

	// --------------------------
	// CONTROLLERS
	// --------------------------
	controllers.InitOrderController(orderService)
	controllers.InitPaymentController(orderService, userService)
	controllers.InitProductController(productService)
	controllers.InitUserController(userRepo) // ✅ FIXED
	controllers.InitCartController(cartService)

	productController := controllers.ProductControllerSingleton()
	adminController := controllers.NewAdminController(productService, orderService, userService)

	// --------------------------
	// ADMIN ROUTES
	// --------------------------
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middlewares.JWTMiddleware(), middlewares.AdminMiddleware())
	{
		adminRoutes.POST("/products", adminController.CreateProduct)
		adminRoutes.PUT("/products/:id", adminController.UpdateProduct)
		adminRoutes.DELETE("/products/:id", adminController.DeleteProduct)

		adminRoutes.GET("/orders", adminController.ListOrders)
		adminRoutes.PATCH("/orders/:id/status", adminController.UpdateOrderStatus)

		adminRoutes.GET("/users", adminController.ListUsers)
		adminRoutes.PATCH("/users/:id", adminController.UpdateUser)
		adminRoutes.DELETE("/users/:id", adminController.DeleteUser)

		adminRoutes.GET("/analytics/sales", adminController.SalesAnalytics)
	}

	// --------------------------
	// PUBLIC PRODUCT ROUTES
	// --------------------------
	r.GET("/products", productController.GetAllProducts)
	r.GET("/products/:id", productController.GetProductByID)
	r.POST("/products", productController.CreateProduct)
	r.PUT("/products/:id", productController.UpdateProduct)
	r.DELETE("/products/:id/image", productController.DeleteProduct)

	// --------------------------
	// AUTH ROUTES
	// --------------------------
	r.POST("/signup", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/test-email", controllers.TestEmail)

	// --------------------------
	// CART ROUTES
	// --------------------------
	cartRoutes := r.Group("/cart")
	cartRoutes.Use(middlewares.JWTMiddleware())
	{
		cartRoutes.POST("", controllers.CreateCart)
		cartRoutes.GET("", controllers.GetCart)
		cartRoutes.PUT("/:id", controllers.UpdateCartItem)
		cartRoutes.DELETE("/:id", controllers.DeleteCartItem)
		cartRoutes.DELETE("", controllers.ClearCart)
	}

	// --------------------------
	// ORDER & PAYMENT ROUTES
	// --------------------------
	orderRoutes := r.Group("/orders")
	orderRoutes.Use(middlewares.JWTMiddleware())
	{
		orderRoutes.POST("", controllers.CreateOrder)
		orderRoutes.GET("", controllers.GetOrders)
		orderRoutes.GET("/:id", controllers.GetOrderByID)
		orderRoutes.PUT("/:id/cancel", controllers.CancelOrder)

		// Payment initialization
		orderRoutes.POST("/:id/pay", controllers.InitializePayment)
	}

	// --------------------------
	// PAYMENT WEBHOOK
	// --------------------------
	r.POST("/payment/webhook", controllers.StripeWebhook)
}
