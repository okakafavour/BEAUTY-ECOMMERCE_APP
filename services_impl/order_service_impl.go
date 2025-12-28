package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"beauty-ecommerce-backend/services"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Ensure at compile-time that orderServiceImpl implements services.OrderService
var _ services.OrderService = (*orderServiceImpl)(nil)

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

// -------------------- CREATE ORDER --------------------
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

		// STOCK CHECK
		if product.Stock <= 0 {
			return order, fmt.Errorf("%s is out of stock", product.Name)
		}
		if item.Quantity > product.Stock {
			return order, fmt.Errorf("only %d unit(s) of %s left in stock", product.Stock, product.Name)
		}

		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price
		subtotal += product.Price * float64(item.Quantity)
	}

	order.Subtotal = subtotal

	// Shipping fee
	switch order.DeliveryType {
	case "standard":
		order.ShippingFee = 3.99
	case "express":
		order.ShippingFee = 4.99
	default:
		order.ShippingFee = 3.99
		order.DeliveryType = "standard"
	}

	order.TotalPrice = order.Subtotal + order.ShippingFee
	order.Status = "pending"
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.CreateOrder(context.Background(), &order); err != nil {
		return order, err
	}

	return order, nil
}

// -------------------- CANCEL ORDER --------------------
func (s *orderServiceImpl) CancelOrder(orderID, userID primitive.ObjectID) (*models.Order, error) {
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
	if err := s.orderRepo.UpdateOrder(order); err != nil {
		return nil, err
	}

	// Restore stock
	if err := s.restoreStock(order); err != nil {
		fmt.Println("⚠️ Failed to restore stock on cancellation:", err)
	}

	return order, nil
}

// -------------------- MARK ORDER AS PAID --------------------
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
	if err := s.orderRepo.UpdateOrder(order); err != nil {
		return err
	}

	// Reduce stock atomically
	for _, item := range order.Items {
		productID, _ := primitive.ObjectIDFromHex(item.ProductID)
		filter := bson.M{"_id": productID, "stock": bson.M{"$gte": item.Quantity}}
		update := bson.M{
			"$inc": bson.M{"stock": -item.Quantity},
			"$set": bson.M{"updated_at": time.Now()},
		}
		res, err := s.productRepo.UpdateWithFilter(filter, update)
		if err != nil || res.MatchedCount == 0 {
			fmt.Printf("⚠️ Not enough stock to decrement for product %s\n", item.ProductName)
		}
	}

	user, _ := s.userRepo.FindById(order.UserID.Hex())
	if err := utils.SendConfirmationEmail(user.Email, user.Name, order.ID.Hex(), order.DeliveryType, order.Subtotal, order.ShippingFee, order.TotalPrice); err != nil {
		fmt.Println("⚠️ Failed to send customer confirmation email:", err)
	}
	if err := utils.SendAdminNotification(order.ID.Hex(), user.Name, user.Email, order.DeliveryType, order.Subtotal, order.ShippingFee, order.TotalPrice); err != nil {
		fmt.Println("⚠️ Failed to send admin notification email:", err)
	}

	return nil
}

// -------------------- PAYMENT FAIL / REFUND / DISPUTE --------------------
func (s *orderServiceImpl) MarkOrderAsFailed(paymentReference string) error {
	return s.handleOrderFailure(paymentReference, "failed")
}

func (s *orderServiceImpl) MarkOrderAsRefunded(paymentReference string) error {
	return s.handleOrderFailure(paymentReference, "refunded")
}

func (s *orderServiceImpl) MarkOrderAsDisputed(paymentReference string) error {
	return s.handleOrderFailure(paymentReference, "disputed")
}

func (s *orderServiceImpl) handleOrderFailure(paymentReference, status string) error {
	order, err := s.orderRepo.FindByReference(paymentReference)
	if err != nil {
		return err
	}

	if order.Status == "pending" || order.Status == "paid" {
		order.Status = status
		order.UpdatedAt = time.Now()
		if err := s.orderRepo.UpdateOrder(order); err != nil {
			return err
		}

		if err := s.restoreStock(order); err != nil {
			fmt.Println("⚠️ Failed to restore stock on order failure:", err)
		}
	}

	return nil
}

// -------------------- RESTORE STOCK --------------------
func (s *orderServiceImpl) restoreStock(order *models.Order) error {
	for _, item := range order.Items {
		productID, _ := primitive.ObjectIDFromHex(item.ProductID)
		update := bson.M{
			"$inc": bson.M{"stock": item.Quantity},
			"$set": bson.M{"updated_at": time.Now()},
		}
		if err := s.productRepo.Update(productID, update); err != nil {
			return err
		}
	}
	return nil
}

// -------------------- SHIPMENT EMAIL --------------------
func (s *orderServiceImpl) SendShippedEmail(order *models.Order) error {
	user, err := s.userRepo.FindById(order.UserID.Hex())
	if err != nil {
		return err
	}

	return utils.SendShipmentEmail(user.Email, user.Name, order.ID.Hex(), order.DeliveryType)
}

// -------------------- OTHER INTERFACE METHODS --------------------
func (s *orderServiceImpl) GetOrdersByUser(userID primitive.ObjectID) ([]models.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

func (s *orderServiceImpl) GetOrderByID(orderID primitive.ObjectID) (*models.Order, error) {
	return s.orderRepo.FindByID(orderID)
}

func (s *orderServiceImpl) GetAllOrders() ([]models.Order, error) {
	return s.orderRepo.FindAll()
}

func (s *orderServiceImpl) UpdateOrderStatus(orderID primitive.ObjectID, status string) error {
	update := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}

	if err := s.orderRepo.Update(orderID, update); err != nil {
		return err
	}

	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}

	if status == "shipped" {
		if err := s.SendShippedEmail(order); err != nil {
			fmt.Println("⚠️ Failed to send shipment email:", err)
		}
	}

	return nil
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

	return map[string]interface{}{
		"total_orders":  len(orders),
		"total_revenue": totalRevenue,
		"status_count":  statusCount,
	}, nil
}

func (s *orderServiceImpl) SaveOrderReference(orderID string, reference string) error {
	return s.orderRepo.UpdateOrderReference(orderID, reference)
}

func (s *orderServiceImpl) GetProductByID(productID primitive.ObjectID) (*models.Product, error) {
	return s.productRepo.FindByID(productID)
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

	pi, err := utils.CreateStripePaymentIntent(order.TotalPrice, "gbp", orderID.Hex(), user.Email)
	if err != nil {
		return "", "", err
	}

	order.PaymentReference = pi.ID
	if err := s.orderRepo.UpdateOrderReference(orderID.Hex(), pi.ID); err != nil {
		return "", "", errors.New("failed to save payment reference")
	}

	return pi.ClientSecret, pi.ID, nil
}
