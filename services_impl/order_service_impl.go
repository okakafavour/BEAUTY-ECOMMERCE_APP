package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type orderServiceImpl struct {
	orderRepo   *repositories.OrderRepository
	productRepo *repositories.ProductRepository
}

// Constructor
func NewOrderService(orderRepo *repositories.OrderRepository, productRepo *repositories.ProductRepository) *orderServiceImpl {
	return &orderServiceImpl{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *orderServiceImpl) CreateOrder(order models.Order) (models.Order, error) {
	var total float64 = 0

	for i, item := range order.Items {
		// Convert string to ObjectID
		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return order, errors.New("invalid product id: " + item.ProductID)
		}

		// Fetch product using FindByID
		product, err := s.productRepo.FindByID(productID)
		if err != nil {
			return order, errors.New("product not found: " + item.ProductID)
		}

		// Populate order item with product details
		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price

		// Add to total
		total += product.Price * float64(item.Quantity)
	}

	order.Total = total
	order.Status = "pending"
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Save order
	err := s.orderRepo.CreateOrder(context.Background(), &order)
	return order, err
}

// // Get all orders for a user
// func (s *orderServiceImpl) GetOrdersByUser(userID string) ([]models.Order, error) {
// 	return s.orderRepo.GetOrdersByUser(context.Background(), userID)
// }

// // Get a single order by ID
// func (s *orderServiceImpl) GetOrderByID(orderID string) (models.Order, error) {
// 	return s.orderRepo.GetOrderByID(context.Background(), orderID)
// }

func (s *orderServiceImpl) GetOrdersByUser(userID primitive.ObjectID) ([]models.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

func (s *orderServiceImpl) GetOrderByID(orderID primitive.ObjectID) (*models.Order, error) {
	return s.orderRepo.FindByID(orderID)
}
