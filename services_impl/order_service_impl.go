package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/utils"
	"context"
	"errors"
	"fmt"
	"os"
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

func (s *orderServiceImpl) CreateOrder(order models.Order) (models.Order, error) {
	if len(order.Items) == 0 {
		return order, errors.New("order must contain at least one item")
	}

	var subtotal float64

	for i, item := range order.Items {
		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return order, errors.New("invalid product ID")
		}

		product, err := s.productRepo.FindByID(productID)
		if err != nil {
			return order, errors.New("product not found")
		}

		if item.Quantity > product.Stock {
			return order, fmt.Errorf("only %d left of %s", product.Stock, product.Name)
		}

		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price
		subtotal += product.Price * float64(item.Quantity)
	}

	order.Subtotal = subtotal

	switch order.DeliveryType {
	case "express":
		order.ShippingFee = 4.99
	default:
		order.DeliveryType = "standard"
		order.ShippingFee = 3.99
	}

	order.TotalPrice = order.Subtotal + order.ShippingFee
	order.Status = "pending"
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.CreateOrder(context.Background(), &order); err != nil {
		return order, err
	}

	// ✅ Email sends correctly now
	go s.notifyUserOrderCreated(&order)
	go s.notifyAdminOrderCreated(&order)

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

	go s.notifyUserPaymentSuccess(order)
	go s.notifyAdminPaymentSuccess(order)

	return nil
}

// -------------------- MARK ORDER AS FAILED --------------------
func (s *orderServiceImpl) MarkOrderAsFailed(paymentReference string) error {
	order, err := s.orderRepo.FindByReference(paymentReference)
	if err != nil {
		return err
	}

	if err := s.restoreStock(order); err != nil {
		fmt.Println("⚠️ Failed to restore stock on order failure:", err)
	}

	go s.notifyUserPaymentFailed(order)
	go s.notifyAdminPaymentFailed(order)

	order.Status = "failed"
	order.UpdatedAt = time.Now()
	return s.orderRepo.UpdateOrder(order)
}

// -------------------- HANDLE REFUND / DISPUTE --------------------
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
func (s *orderServiceImpl) SendShippedEmail(order *models.Order) {
	user, err := s.userRepo.FindById(order.UserID.Hex())
	if err != nil {
		fmt.Println("⚠️ Could not find user for shipment email:", err)
		return
	}
	utils.SendShipmentEmail(user.Email, user.Name, order.ID.Hex(), order.DeliveryType)
}

// -------------------- NOTIFICATIONS --------------------
func (s *orderServiceImpl) notifyAdminOrderCreated(order *models.Order) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return
	}
	subject := fmt.Sprintf("🛒 New Order Created - %s", order.ID.Hex())
	itemsHTML := ""
	for _, item := range order.Items {
		itemsHTML += fmt.Sprintf("<li>%s × %d — £%.2f</li>", item.ProductName, item.Quantity, item.Price*float64(item.Quantity))
	}
	html := fmt.Sprintf(`
		<h2>New Order Created</h2>
		<p><strong>Customer:</strong> %s (%s)</p>
		<p><strong>Order ID:</strong> %s</p>
		<p><strong>Delivery:</strong> %s</p>
		<h3>Items</h3><ul>%s</ul>
		<p><strong>Subtotal:</strong> £%.2f</p>
		<p><strong>Shipping:</strong> £%.2f</p>
		<p><strong>Total:</strong> £%.2f</p>
		<p>Status: <b>Pending payment</b></p>
	`, order.CustomerName, order.CustomerEmail, order.ID.Hex(), order.DeliveryType, itemsHTML, order.Subtotal, order.ShippingFee, order.TotalPrice)
	utils.QueueEmail(adminEmail, "Admin", subject, html)
}

func (s *orderServiceImpl) notifyUserOrderCreated(order *models.Order) {
	user, err := s.userRepo.FindById(order.UserID.Hex())
	if err != nil {
		fmt.Println("⚠️ Could not find user for order email:", err)
		return
	}
	utils.SendConfirmationEmail(user.Email, user.Name, order.ID.Hex(), order.DeliveryType, order.Subtotal, order.ShippingFee, order.TotalPrice)
}

func (s *orderServiceImpl) notifyUserPaymentSuccess(order *models.Order) {
	user, err := s.userRepo.FindById(order.UserID.Hex())
	if err != nil {
		fmt.Println("⚠️ Could not find user for payment success email:", err)
		return
	}
	utils.SendConfirmationEmail(user.Email, user.Name, order.ID.Hex(), order.DeliveryType, order.Subtotal, order.ShippingFee, order.TotalPrice)
}

func (s *orderServiceImpl) notifyAdminPaymentSuccess(order *models.Order) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return
	}
	user, _ := s.userRepo.FindById(order.UserID.Hex())
	subject := fmt.Sprintf("Order Paid - %s", order.ID.Hex())
	html := fmt.Sprintf(`<p>Order <b>%s</b> paid by <b>%s</b> (%s).</p>
		<p>Delivery: %s</p>
		<p>Subtotal: £%.2f | Shipping: £%.2f | Total: £%.2f</p>`,
		order.ID.Hex(), user.Name, user.Email, order.DeliveryType, order.Subtotal, order.ShippingFee, order.TotalPrice)
	utils.QueueEmail(adminEmail, "Admin", subject, html)
}

func (s *orderServiceImpl) notifyUserPaymentFailed(order *models.Order) {
	user, err := s.userRepo.FindById(order.UserID.Hex())
	if err != nil {
		fmt.Println("⚠️ Could not find user for payment failed email:", err)
		return
	}
	utils.SendFailedPaymentEmail(user.Email, user.Name, order.ID.Hex())
}

func (s *orderServiceImpl) notifyAdminPaymentFailed(order *models.Order) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return
	}
	user, _ := s.userRepo.FindById(order.UserID.Hex())
	subject := fmt.Sprintf("Payment FAILED - %s", order.ID.Hex())
	html := fmt.Sprintf("<p>Payment for order <b>%s</b> by <b>%s</b> (%s) FAILED.</p>", order.ID.Hex(), user.Name, user.Email)
	utils.QueueEmail(adminEmail, "Admin", subject, html)
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

	if err := s.restoreStock(order); err != nil {
		fmt.Println("⚠️ Failed to restore stock on cancellation:", err)
	}

	return order, nil
}

// -------------------- GET ORDERS --------------------
func (s *orderServiceImpl) GetAllOrders() ([]models.Order, error) {
	return s.orderRepo.FindAll()
}

func (s *orderServiceImpl) GetOrdersByUser(userID primitive.ObjectID) ([]models.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

func (s *orderServiceImpl) GetOrderByID(orderID primitive.ObjectID) (*models.Order, error) {
	return s.orderRepo.FindByID(orderID)
}

// -------------------- UPDATE ORDER STATUS --------------------
func (s *orderServiceImpl) UpdateOrderStatus(orderID primitive.ObjectID, status string) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	order.Status = status
	order.UpdatedAt = time.Now()
	if err := s.orderRepo.UpdateOrder(order); err != nil {
		return err
	}

	if status == "shipped" {
		go s.SendShippedEmail(order)
	}

	return nil
}

// -------------------- INITIALIZE PAYMENT --------------------
func (s *orderServiceImpl) InitializePayment(orderID, userID primitive.ObjectID, email string) (string, string, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return "", "", errors.New("order not found")
	}
	if order.UserID != userID {
		return "", "", errors.New("unauthorized")
	}

	reference := primitive.NewObjectID().Hex()
	if err := s.SaveOrderReference(orderID.Hex(), reference); err != nil {
		return "", "", err
	}

	paymentURL := fmt.Sprintf("https://payment-provider.com/pay/%s", reference)
	return reference, paymentURL, nil
}

// -------------------- SAVE ORDER REFERENCE --------------------
func (s *orderServiceImpl) SaveOrderReference(orderID string, reference string) error {
	return s.orderRepo.UpdateOrderReference(orderID, reference)
}

// -------------------- GET PRODUCT --------------------
func (s *orderServiceImpl) GetProductByID(productID primitive.ObjectID) (*models.Product, error) {
	return s.productRepo.FindByID(productID)
}

func (s *orderServiceImpl) GetSalesAnalytics() (map[string]interface{}, error) {
	// Example: return empty analytics
	return map[string]interface{}{}, nil
}
