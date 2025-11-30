package repositories

import (
	"beauty-ecommerce-backend/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WishlistRepository struct {
	Collection *mongo.Collection
}

func NewWishlistRepository(collection *mongo.Collection) *WishlistRepository {
	return &WishlistRepository{Collection: collection}
}

// Find wishlist by user
func (r *WishlistRepository) FindByUser(userID primitive.ObjectID) (*models.Wishlist, error) {
	var wishlist models.Wishlist
	err := r.Collection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&wishlist)
	if err != nil {
		return nil, err
	}
	return &wishlist, nil
}

// Create wishlist
func (r *WishlistRepository) Create(wishlist *models.Wishlist) error {
	wishlist.CreatedAt = time.Now()
	wishlist.UpdatedAt = time.Now()
	_, err := r.Collection.InsertOne(context.TODO(), wishlist)
	return err
}

// Update wishlist products
func (r *WishlistRepository) UpdateProducts(userID primitive.ObjectID, products []primitive.ObjectID) error {
	_, err := r.Collection.UpdateOne(
		context.TODO(),
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"product_ids": products, "updated_at": time.Now()}},
	)
	return err
}
