package repositories

import (
	"context"
	"fmt"
	"time"

	"beauty-ecommerce-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
	}
}

// --------------------------
// CREATE
// --------------------------
func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, order)
	return err
}

// --------------------------
// FIND BY ID
// --------------------------
func (r *OrderRepository) FindByID(orderID primitive.ObjectID) (*models.Order, error) {
	var order models.Order
	err := r.collection.FindOne(context.Background(), bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// --------------------------
// FIND BY USER
// --------------------------
func (r *OrderRepository) FindByUserID(userID primitive.ObjectID) ([]models.Order, error) {
	ctx := context.Background()
	filter := bson.M{"user_id": userID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

// --------------------------
// FIND BY PAYMENT REFERENCE
// --------------------------
func (r *OrderRepository) FindByReference(reference string) (*models.Order, error) {
	var order models.Order
	err := r.collection.FindOne(context.Background(), bson.M{"payment_reference": reference}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// --------------------------
// UPDATE ENTIRE ORDER
// --------------------------
func (r *OrderRepository) UpdateOrder(order *models.Order) error {
	_, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": order.ID},
		bson.M{"$set": order},
	)
	return err
}

// --------------------------
// UPDATE PAYMENT REFERENCE
// --------------------------
func (r *OrderRepository) UpdateOrderReference(orderID, reference string) error {
	id, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"payment_reference": reference,
			"status":            "pending",
			"updated_at":        time.Now(),
		},
	}
	res, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("no order found to update reference")
	}
	return nil
}

// --------------------------
// ADMIN: FIND ALL ORDERS
// --------------------------
func (r *OrderRepository) FindAll() ([]models.Order, error) {
	ctx := context.Background()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

// --------------------------
// ADMIN: UPDATE ORDER STATUS
func (r *OrderRepository) Update(orderID primitive.ObjectID, update bson.M) error {
	filter := bson.M{"_id": orderID}
	updateBson := bson.M{"$set": update}

	res, err := r.collection.UpdateOne(context.Background(), filter, updateBson)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("no order found to update")
	}
	return nil
}

func (r *OrderRepository) MarkFailed(paymentReference string) error {
	filter := bson.M{"payment_reference": paymentReference}
	update := bson.M{"$set": bson.M{
		"status":     "failed",
		"updated_at": time.Now(),
	}}

	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *OrderRepository) MarkPaid(paymentReference string) error {
	filter := bson.M{"payment_reference": paymentReference}

	update := bson.M{"$set": bson.M{
		"status":     "paid",
		"updated_at": time.Now(),
	}}

	res, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("no order found to mark as paid")
	}

	return nil
}

func (r *OrderRepository) MarkRefunded(paymentReference string) error {
	filter := bson.M{"payment_reference": paymentReference}
	update := bson.M{
		"$set": bson.M{
			"status":     "refunded",
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *OrderRepository) MarkDisputed(paymentReference string) error {
	filter := bson.M{"payment_reference": paymentReference}
	update := bson.M{
		"$set": bson.M{
			"status":     "disputed",
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}
