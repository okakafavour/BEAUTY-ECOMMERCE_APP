package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type orderServiceImpl struct {
	orderRepo   *repositories.OrderRepository
	productRepo *repositories.ProductRepository
	userRepo    *repositories.UserRepository
}

// Constructor
func NewOrderService(orderRepo *repositories.OrderRepository, productRepo *repositories.ProductRepository, userRepo *repositories.UserRepository) *orderServiceImpl {
	return &orderServiceImpl{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
	}
}

func (s *orderServiceImpl) CreateOrder(order models.Order) (models.Order, error) {
	if len(order.Items) == 0 {
		return order, errors.New("order must contain at least one item")
	}

	var subtotal float64

	for i, item := range order.Items {
		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return order, fmt.Errorf("invalid product ID: %s", item.ProductID)
		}

		product, err := s.productRepo.FindByID(productID)
		if err != nil {
			return order, fmt.Errorf("product not found: %s", item.ProductID)
		}

		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price

		subtotal += product.Price * float64(item.Quantity)
	}

	order.Subtotal = subtotal
	order.ShippingFee = 0
	order.TotalPrice = subtotal + order.ShippingFee

	order.Status = "pending"
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	err := s.orderRepo.CreateOrder(context.Background(), &order)
	if err != nil {
		return order, err
	}

	return order, nil
}

func (s *orderServiceImpl) GetOrdersByUser(userID primitive.ObjectID) ([]models.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

func (s *orderServiceImpl) GetOrderByID(orderID primitive.ObjectID) (*models.Order, error) {
	return s.orderRepo.FindByID(orderID)
}

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

func (s *orderServiceImpl) InitializePayment(orderID, userID primitive.ObjectID, email string) (string, string, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return "", "", errors.New("order not found")
	}

	if order.UserID != userID {
		return "", "", errors.New("you are not allowed to pay for this order")
	}

	user, err := s.userRepo.FindById(userID.Hex())
	if err != nil {
		return "", "", errors.New("user not found")
	}

	pi, err := utils.CreateStripePaymentIntent(order.TotalPrice, "ngn", orderID.Hex(), user.Email)
	if err != nil {
		return "", "", err
	}

	order.PaymentReference = pi.ID
	if err := s.orderRepo.UpdateOrderReference(orderID.Hex(), pi.ID); err != nil {
		return "", "", errors.New("failed to save payment reference")
	}

	return pi.ClientSecret, pi.ID, nil
}

func (s *orderServiceImpl) MarkOrderAsPaid(paymentReference string) error {
	order, err := s.orderRepo.FindByReference(paymentReference)
	if err != nil {
		return errors.New("order not found for this payment reference")
	}

	if order.Status == "paid" {
		return nil
	}

	if order.Status != "pending" {
		return errors.New("order cannot be marked as paid")
	}

	order.Status = "paid"
	order.UpdatedAt = time.Now()
	return s.orderRepo.UpdateOrder(order)
}

func (s *orderServiceImpl) SaveOrderReference(orderID string, reference string) error {
	return s.orderRepo.UpdateOrderReference(orderID, reference)
}

func (s *orderServiceImpl) GetAllOrders() ([]models.Order, error) {
	return s.orderRepo.FindAll()
}

func (s *orderServiceImpl) UpdateOrderStatus(orderID primitive.ObjectID, status string) error {
	update := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}
	return s.orderRepo.Update(orderID, update)
}

func (s *orderServiceImpl) GetSalesAnalytics() (map[string]interface{}, error) {
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		return nil, err
	}

	totalRevenue := 0.0
	statusCount := map[string]int{}
	for _, o := range orders {
		totalRevenue += o.TotalPrice
		statusCount[o.Status]++
	}

	data := map[string]interface{}{
		"total_orders":  len(orders),
		"total_revenue": totalRevenue,
		"status_count":  statusCount,
	}

	return data, nil
}

// Payment webhook helpers
func (s *orderServiceImpl) MarkOrderAsFailed(paymentReference string) error {
	return s.orderRepo.MarkFailed(paymentReference)
}

func (s *orderServiceImpl) MarkOrderAsRefunded(paymentReference string) error {
	return s.orderRepo.MarkRefunded(paymentReference)
}

func (s *orderServiceImpl) MarkOrderAsDisputed(paymentReference string) error {
	return s.orderRepo.MarkDisputed(paymentReference)
}
