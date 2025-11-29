package services

import (
	"beauty-ecommerce-backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService interface {
	CreateOrder(order models.Order) (models.Order, error)
	GetOrdersByUser(userID primitive.ObjectID) ([]models.Order, error)
	GetOrderByID(orderID primitive.ObjectID) (*models.Order, error)
}
