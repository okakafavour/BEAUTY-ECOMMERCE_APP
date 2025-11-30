package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/utils"
	"context"
	"errors"
	"fmt"
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

// CreateOrder calculates total, populates items, and saves the order
func (s *orderServiceImpl) CreateOrder(order models.Order) (models.Order, error) {
	var total float64

	for i, item := range order.Items {
		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return order, errors.New("invalid product id: " + item.ProductID)
		}

		product, err := s.productRepo.FindByID(productID)
		if err != nil {
			return order, errors.New("product not found: " + item.ProductID)
		}

		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price
		total += product.Price * float64(item.Quantity)
	}

	order.Total = total
	order.Status = "pending"
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	err := s.orderRepo.CreateOrder(context.Background(), &order)
	return order, err
}

// GetOrdersByUser fetches all orders for a user
func (s *orderServiceImpl) GetOrdersByUser(userID primitive.ObjectID) ([]models.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

// GetOrderByID fetches a single order by ID
func (s *orderServiceImpl) GetOrderByID(orderID primitive.ObjectID) (*models.Order, error) {
	return s.orderRepo.FindByID(orderID)
}

// CancelOrder cancels a pending order if owned by the user
func (s *orderServiceImpl) CancelOrder(orderID primitive.ObjectID, userID primitive.ObjectID) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.UserID != userID {
		return nil, errors.New("you cannot cancel this order")
	}

	if order.Status != "pending" {
		return nil, errors.New("order cannot be cancelled")
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()
	err = s.orderRepo.UpdateOrder(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// InitializePayment generates a reference and calls Paystack
func (s *orderServiceImpl) InitializePayment(orderID, userID primitive.ObjectID, email string) (string, string, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return "", "", errors.New("order not found")
	}

	if order.UserID != userID {
		return "", "", errors.New("you are not allowed to pay for this order")
	}

	reference := fmt.Sprintf("PSK_REF_%s", orderID.Hex())

	res, err := utils.PaystackInitialize(email, order.Total, reference)
	if err != nil {
		return "", "", err
	}

	// Save the reference in DB
	order.PaymentReference = reference
	s.orderRepo.UpdateOrder(order)

	return res.Data.AuthorizationURL, reference, nil
}

// MarkOrderAsPaid sets order status to 'paid' after verification
func (s *orderServiceImpl) MarkOrderAsPaid(reference string) error {
	order, err := s.orderRepo.FindByReference(reference)
	if err != nil {
		return errors.New("order not found for this reference")
	}

	if order.Status == "paid" {
		return nil // already paid, no error
	}

	if order.Status != "pending" {
		return errors.New("order cannot be marked as paid")
	}

	order.Status = "paid"
	order.UpdatedAt = time.Now()
	return s.orderRepo.UpdateOrder(order)
}

// SaveOrderReference saves a generated payment reference before initializing Paystack payment
func (s *orderServiceImpl) SaveOrderReference(orderID string, reference string) error {
	return s.orderRepo.UpdateOrderReference(orderID, reference)
}
