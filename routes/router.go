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
	// SERVICES
	// --------------------------
	productRepo := repositories.NewProductRepository(db)
	productService := servicesimpl.NewProductService(productRepo)

	orderRepo := repositories.NewOrderRepository(db)
	orderService := servicesimpl.NewOrderService(orderRepo, productRepo)

	userService := servicesimpl.NewUserService()

	// --------------------------
	// ADMIN CONTROLLER
	// --------------------------
	adminController := controllers.NewAdminController(productService, orderService, userService)

	// --------------------------
	// ADMIN ROUTES
	// --------------------------
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middlewares.JWTMiddleware(), middlewares.AdminMiddleware())
	{
		// Products
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

		// Analytics (optional)
		adminRoutes.GET("/analytics/sales", adminController.SalesAnalytics)
	}

	/// --------------------------
	// PUBLIC PRODUCT ROUTES
	// --------------------------
	controllers.InitProductController(productService)
	productController := controllers.ProductControllerSingleton()

	r.GET("/products", productController.GetAllProducts)
	r.GET("/products/:id", productController.GetProductByID)
	r.POST("/products", productController.CreateProduct)    // JSON-based creation
	r.PUT("/products/:id", productController.UpdateProduct) // JSON-based update

	// --------------------------
	// AUTH ROUTES
	// --------------------------
	controllers.InitUserController()
	r.POST("/signup", controllers.Register)
	r.POST("/login", controllers.Login)

	// --------------------------
	// CART ROUTES
	// --------------------------
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

	// --------------------------
	// ORDER & PAYMENT ROUTES
	// --------------------------
	controllers.InitOrderController(orderService)
	controllers.InitPaymentController(orderService)

	orderRoutes := r.Group("/orders")
	orderRoutes.Use(middlewares.JWTMiddleware())
	{
		orderRoutes.POST("", controllers.CreateOrder)
		orderRoutes.GET("", controllers.GetOrders)
		orderRoutes.GET("/:id", controllers.GetOrderByID)
		orderRoutes.PUT("/:id/cancel", controllers.CancelOrder)
		orderRoutes.POST("/:id/pay", controllers.InitializePayment)
	}

	// --------------------------
	// PAYMENT CALLBACK
	// --------------------------
	// r.GET("/payment/success", controllers.PaymentSuccess)

	// --------------------------
	// REVIEW ROUTES
	// --------------------------
	reviewRepo := repositories.NewReviewRepository(db)
	reviewService := services.NewReviewService(reviewRepo /*, orderRepo optional*/)
	reviewController := controllers.NewReviewController(reviewService)

	reviewRoutes := r.Group("/reviews")
	{
		reviewRoutes.GET("/:productId", reviewController.GetProductReviews)

		reviewRoutesAuth := reviewRoutes.Group("")
		reviewRoutesAuth.Use(middlewares.JWTMiddleware())
		{
			reviewRoutesAuth.POST("", reviewController.CreateReview)
			reviewRoutesAuth.PUT("/:id", reviewController.UpdateReview)
			reviewRoutesAuth.DELETE("/:id", reviewController.DeleteReview)
		}
	}

	// --------------------------
	// WISHLIST ROUTES
	// --------------------------
	wishlistRepo := repositories.NewWishlistRepository(db.Collection("wishlists"))
	wishlistService := servicesimpl.NewWishlistService(wishlistRepo)
	wishlistController := controllers.NewWishlistController(wishlistService)

	wishlistRoutes := r.Group("/wishlist")
	wishlistRoutes.Use(middlewares.JWTMiddleware())
	{
		wishlistRoutes.GET("", wishlistController.GetWishlist)
		wishlistRoutes.POST("/add", wishlistController.AddToWishlist)
		wishlistRoutes.POST("/remove", wishlistController.RemoveFromWishlist)
	}
}
