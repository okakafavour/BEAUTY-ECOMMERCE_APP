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

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, order)
	return err
}

func (r *OrderRepository) FindByID(orderID primitive.ObjectID) (*models.Order, error) {
	var order models.Order
	err := r.collection.FindOne(context.Background(), bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByReference(reference string) (*models.Order, error) {
	var order models.Order
	err := r.collection.FindOne(context.Background(), bson.M{"payment_reference": reference}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateOrder(order *models.Order) error {
	_, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": order.ID},
		bson.M{"$set": order},
	)
	return err
}

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

// In repositories/order_repository.go
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
