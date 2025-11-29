package repositories

import (
	"beauty-ecommerce-backend/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartRepository struct {
	collection *mongo.Collection
}

func NewCartRepository(db *mongo.Database) *CartRepository {
	return &CartRepository{collection: db.Collection("cart")}
}

func (r *CartRepository) AddToCart(ctx context.Context, item *models.CartItem) error {
	_, err := r.collection.InsertOne(ctx, item)
	return err
}

func (r *CartRepository) GetUserCart(ctx context.Context, userID primitive.ObjectID) ([]models.CartItem, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.CartItem
	for cursor.Next(ctx) {
		var item models.CartItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *CartRepository) UpdateQuantity(ctx context.Context, cartItemID primitive.ObjectID, quantity int) error {
	_, err := r.collection.UpdateByID(ctx, cartItemID, bson.M{
		"$set": bson.M{"quantity": quantity, "updated_at": time.Now()},
	})
	return err
}

func (r *CartRepository) DeleteCartItem(ctx context.Context, cartItemID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": cartItemID})
	return err
}

func (r *CartRepository) FindByID(id primitive.ObjectID) (*models.CartItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cartItem models.CartItem
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&cartItem)
	if err != nil {
		return nil, err
	}
	return &cartItem, nil
}
