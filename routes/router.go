package routes

import (
	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/controllers"
	"beauty-ecommerce-backend/middlewares"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
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
	reviewRepo := repositories.NewReviewRepository(db)

	// --------------------------
	// SERVICES
	// --------------------------
	productService := servicesimpl.NewProductService(productRepo)
	userService := servicesimpl.NewUserService(userRepo)
	orderService := servicesimpl.NewOrderService(orderRepo, productRepo, userRepo)
	cartService := servicesimpl.NewCartService(cartRepo)
	reviewService := services.NewReviewService(reviewRepo)

	// --------------------------
	// CONTROLLERS
	// --------------------------
	controllers.InitOrderController(orderService)
	controllers.InitPaymentController(orderService, userService)
	controllers.InitProductController(productService)
	controllers.InitUserController(userRepo)
	controllers.InitCartController(cartService)

	productController := controllers.ProductControllerSingleton()
	adminController := controllers.NewAdminController(productService, orderService, userService)
	reviewController := controllers.NewReviewController(reviewService)

	// ============================================================
	// ADMIN ROUTES  (JWT + ADMIN REQUIRED)
	// ============================================================
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middlewares.JWTMiddleware(), middlewares.AdminMiddleware())
	{
		// Product management
		adminRoutes.POST("/products", adminController.CreateProduct)
		adminRoutes.PUT("/products/:id", adminController.UpdateProduct)
		adminRoutes.DELETE("/products/:id", adminController.DeleteProduct)

		// Orders
		adminRoutes.GET("/orders", adminController.ListOrders)
		adminRoutes.PATCH("/orders/:id/status", adminController.UpdateOrderStatus)

		// Users
		adminRoutes.GET("/users", adminController.ListUsers)
		adminRoutes.PATCH("/users/:id", adminController.UpdateUser)
		adminRoutes.DELETE("/users/:id", adminController.DeleteUser)

		// Analytics
		adminRoutes.GET("/analytics/sales", adminController.SalesAnalytics)
	}

	// ============================================================
	// PUBLIC PRODUCT ROUTES (Safe)
	// ============================================================
	r.GET("/products", productController.GetAllProducts)
	r.GET("/products/:id", productController.GetProductByID)

	// ============================================================
	// REVIEWS
	// ============================================================
	reviewRoutes := r.Group("/reviews")
	reviewRoutes.Use(middlewares.JWTMiddleware())
	{
		reviewRoutes.POST("", reviewController.CreateReview)
		reviewRoutes.PUT("/:id", reviewController.UpdateReview)
		reviewRoutes.DELETE("/:id", reviewController.DeleteReview)
	}
	r.GET("/products/:id/reviews", reviewController.GetProductReviews)

	// ============================================================
	// AUTH
	// ============================================================
	r.POST("/signup", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/test-email", controllers.TestEmail)

	// ============================================================
	// CART ROUTES
	// ============================================================
	cartRoutes := r.Group("/cart")
	cartRoutes.Use(middlewares.JWTMiddleware())
	{
		cartRoutes.POST("", controllers.CreateCart)
		cartRoutes.GET("", controllers.GetCart)
		cartRoutes.PUT("/:id", controllers.UpdateCartItem)
		cartRoutes.DELETE("/:id", controllers.DeleteCartItem)
		cartRoutes.DELETE("", controllers.ClearCart)
	}

	// ============================================================
	// ORDERS + PAYMENTS
	// ============================================================
	orderRoutes := r.Group("/orders")
	orderRoutes.Use(middlewares.JWTMiddleware())
	{
		orderRoutes.POST("", controllers.CreateOrder)
		orderRoutes.GET("", controllers.GetOrders)
		orderRoutes.GET("/:id", controllers.GetOrderByID)
		orderRoutes.PUT("/:id/cancel", controllers.CancelOrder)

		orderRoutes.POST("/:id/pay", controllers.InitializePayment)
	}

}
