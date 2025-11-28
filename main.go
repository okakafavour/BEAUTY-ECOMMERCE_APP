package main

import (
	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.ConnectDB() // sets global DB
	router := gin.Default()
	routes.SetUpRoutes(router) // no second argument needed
	router.Run(":8080")
}
