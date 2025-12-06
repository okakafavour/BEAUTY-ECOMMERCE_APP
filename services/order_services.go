package services

import (
	"beauty-ecommerce-backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService interface {
	// User operations (existing)
	CreateOrder(order models.Order) (models.Order, error)
	GetOrdersByUser(userID primitive.ObjectID) ([]models.Order, error)
	GetOrderByID(orderID primitive.ObjectID) (*models.Order, error)
	CancelOrder(orderID primitive.ObjectID, userID primitive.ObjectID) (*models.Order, error)
	InitializePayment(orderID, userID primitive.ObjectID, email string) (string, string, error)
	MarkOrderAsPaid(reference string) error
	SaveOrderReference(orderID string, reference string) error
	MarkOrderAsFailed(paymentReference string) error

	// Admin operations (new)
	GetAllOrders() ([]models.Order, error)
	UpdateOrderStatus(orderID primitive.ObjectID, status string) error
	GetSalesAnalytics() (map[string]interface{}, error) // optional
	MarkOrderAsRefunded(paymentReference string) error
	MarkOrderAsDisputed(paymentReference string) error
}
