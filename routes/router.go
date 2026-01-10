package routes

// import (
// 	"beauty-ecommerce-backend/config"
// 	"beauty-ecommerce-backend/controllers"
// 	"beauty-ecommerce-backend/middlewares"
// 	"beauty-ecommerce-backend/repositories"
// 	"beauty-ecommerce-backend/services"
// 	servicesimpl "beauty-ecommerce-backend/services_impl"
// 	"os"

// 	"github.com/gin-gonic/gin"
// )

// func SetUpRoutes(r *gin.Engine) {
// 	db := config.DB

// 	// --------------------------
// 	// REPOSITORIES
// 	// --------------------------
// 	userRepo := repositories.NewUserRepository(db)
// 	productRepo := repositories.NewProductRepository(db)
// 	orderRepo := repositories.NewOrderRepository(db)
// 	cartRepo := repositories.NewCartRepository(db)
// 	reviewRepo := repositories.NewReviewRepository(db)
// 	wishlistCollection := db.Collection("wishlists")
// 	wishlistRepo := repositories.NewWishlistRepository(wishlistCollection)

// 	// --------------------------
// 	// SERVICES
// 	// --------------------------
// 	userService := servicesimpl.NewUserService(userRepo)
// 	productService := servicesimpl.NewProductService(productRepo)
// 	orderService := servicesimpl.NewOrderService(orderRepo, productRepo, userRepo)
// 	cartService := servicesimpl.NewCartService(cartRepo)
// 	reviewService := services.NewReviewService(reviewRepo) // stays as `services`
// 	wishlistService := servicesimpl.NewWishlistService(wishlistRepo, productService)

// 	// --------------------------
// 	// CONTROLLERS
// 	// --------------------------
// 	controllers.InitUserController(userRepo)
// 	controllers.InitOrderController(orderService)
// 	controllers.InitPaymentController(orderService, userService)
// 	controllers.InitProductController(productService)
// 	controllers.InitCartController(cartService)

// 	productController := controllers.ProductControllerSingleton()
// 	adminController := controllers.NewAdminController(productService, orderService, userService)
// 	adminAuthController := controllers.NewAdminAuthController()
// 	reviewController := controllers.NewReviewController(reviewService)
// 	wishlistController := controllers.NewWishlistController(wishlistService)

// 	// --------------------------
// 	// ROUTES
// 	// --------------------------

// 	// ADMIN AUTH
// 	r.POST("/admin/login", adminAuthController.AdminLogin)

// 	// ADMIN (JWT + ADMIN)
// 	adminRoutes := r.Group("/admin")
// 	adminRoutes.Use(middlewares.JWTMiddleware(), middlewares.AdminMiddleware())
// 	{
// 		adminRoutes.POST("/products", adminController.CreateProduct)
// 		adminRoutes.PUT("/products/:id", adminController.UpdateProduct)
// 		adminRoutes.DELETE("/products/:id", adminController.DeleteProduct)

// 		adminRoutes.GET("/orders", adminController.ListOrders)
// 		adminRoutes.PATCH("/orders/:id/status", adminController.UpdateOrderStatus)

// 		adminRoutes.GET("/users", adminController.ListUsers)
// 		adminRoutes.PATCH("/users/:id", adminController.UpdateUser)
// 		adminRoutes.DELETE("/users/:id", adminController.DeleteUser)

// 		adminRoutes.GET("/analytics/sales", adminController.SalesAnalytics)
// 	}

// 	// PUBLIC PRODUCTS
// 	r.GET("/products", productController.GetAllProducts)
// 	r.GET("/products/:id", productController.GetProductByID)

// 	// REVIEWS
// 	reviewRoutes := r.Group("/reviews")
// 	reviewRoutes.Use(middlewares.JWTMiddleware())
// 	{
// 		reviewRoutes.POST("", reviewController.CreateReview)
// 		reviewRoutes.PUT("/:id", reviewController.UpdateReview)
// 		reviewRoutes.DELETE("/:id", reviewController.DeleteReview)
// 	}
// 	r.GET("/products/:id/reviews", reviewController.GetProductReviews)

// 	// AUTH
// 	r.POST("/signup", controllers.Register)
// 	r.POST("/login", controllers.Login)
// 	r.POST("/auth/forgot-password", controllers.ForgotPassword)
// 	r.GET("/reset-password", controllers.ResetPassword)
// 	r.POST("/auth/reset-password", controllers.ResetPassword)
// 	// r.GET("/test-email", controllers.TestEmail)

// 	// CART
// 	cartRoutes := r.Group("/cart")
// 	cartRoutes.Use(middlewares.JWTMiddleware())
// 	{
// 		cartRoutes.POST("", controllers.CreateCart)
// 		cartRoutes.GET("", controllers.GetCart)
// 		cartRoutes.PUT("/:id", controllers.UpdateCartItem)
// 		cartRoutes.DELETE("/:id", controllers.DeleteCartItem)
// 		cartRoutes.DELETE("", controllers.ClearCart)
// 	}

// 	// ORDERS + PAYMENTS
// 	orderRoutes := r.Group("/orders")
// 	orderRoutes.Use(middlewares.JWTMiddleware())
// 	{
// 		orderRoutes.POST("", controllers.CreateOrder)
// 		orderRoutes.GET("", controllers.GetOrders)
// 		orderRoutes.GET("/:id", controllers.GetOrderByID)
// 		orderRoutes.PUT("/:id/cancel", controllers.CancelOrder)
// 		orderRoutes.POST("/:id/pay", controllers.InitializePayment)
// 	}

// 	// WISHLIST
// 	wishlistRoutes := r.Group("/wishlist")
// 	wishlistRoutes.Use(middlewares.JWTMiddleware())
// 	{
// 		wishlistRoutes.GET("", wishlistController.GetWishlistPaginated)
// 		wishlistRoutes.POST("/add", wishlistController.AddToWishlist)
// 		wishlistRoutes.POST("/remove", wishlistController.RemoveFromWishlist)
// 	}

// 	// USER PROFILE
// 	userRoutes := r.Group("/users")
// 	userRoutes.Use(middlewares.JWTMiddleware())
// 	{
// 		userRoutes.GET("/me", controllers.GetProfile)
// 	}

// 	// VERSION
// 	r.GET("/version", func(c *gin.Context) {
// 		c.JSON(200, gin.H{"version": "wishlist_update_2025-12-19"})
// 	})

// 	r.GET("/test-env", func(c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"SMTP_HOST":     os.Getenv("SMTP_HOST"),
// 			"SMTP_PORT":     os.Getenv("SMTP_PORT"),
// 			"SMTP_USER":     os.Getenv("SMTP_USERNAME"),
// 			"SMTP_PASS_SET": os.Getenv("SMTP_PASSWORD") != "",
// 			"SMTP_FROM":     os.Getenv("SMTP_FROM"),
// 		})
// 	})

// 	r.GET("/test-proof-email", controllers.SendProofEmail)

// }
