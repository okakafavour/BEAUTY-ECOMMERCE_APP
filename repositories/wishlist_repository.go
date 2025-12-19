package repositories

import (
	"beauty-ecommerce-backend/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *WishlistRepository) AddProduct(
	userID primitive.ObjectID,
	productID primitive.ObjectID,
) error {

	filter := bson.M{"user_id": userID}

	update := bson.M{
		"$addToSet": bson.M{
			"product_ids": productID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"user_id":    userID,
			"created_at": time.Now(),
		},
	}

	_, err := r.Collection.UpdateOne(
		context.TODO(),
		filter,
		update,
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *WishlistRepository) RemoveProduct(
	userID primitive.ObjectID,
	productID primitive.ObjectID,
) error {

	_, err := r.Collection.UpdateOne(
		context.TODO(),
		bson.M{"user_id": userID},
		bson.M{
			"$pull": bson.M{
				"product_ids": productID,
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

func (r *WishlistRepository) GetPaginated(
	userID primitive.ObjectID,
	offset, limit int,
) ([]primitive.ObjectID, int64, error) {

	var wishlist models.Wishlist

	err := r.Collection.FindOne(
		context.TODO(),
		bson.M{"user_id": userID},
	).Decode(&wishlist)

	if err == mongo.ErrNoDocuments {
		return []primitive.ObjectID{}, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(wishlist.ProductIDs))

	if offset > len(wishlist.ProductIDs) {
		return []primitive.ObjectID{}, total, nil
	}

	end := offset + limit
	if end > len(wishlist.ProductIDs) {
		end = len(wishlist.ProductIDs)
	}

	return wishlist.ProductIDs[offset:end], total, nil
}
