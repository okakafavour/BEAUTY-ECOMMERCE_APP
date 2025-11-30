package repositories

import (
	"beauty-ecommerce-backend/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewRepository struct {
	Collection *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) *ReviewRepository {
	return &ReviewRepository{
		Collection: db.Collection("reviews"),
	}
}

func (r *ReviewRepository) Create(review *models.Review) error {
	_, err := r.Collection.InsertOne(context.Background(), review)
	return err
}

func (r *ReviewRepository) FindByID(id primitive.ObjectID) (*models.Review, error) {
	var review models.Review
	err := r.Collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&review)
	return &review, err
}

func (r *ReviewRepository) Update(id primitive.ObjectID, update bson.M) error {
	_, err := r.Collection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	return err
}

func (r *ReviewRepository) Delete(id primitive.ObjectID) error {
	_, err := r.Collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

func (r *ReviewRepository) GetByProduct(productID primitive.ObjectID) ([]models.Review, error) {
	cursor, err := r.Collection.Find(context.Background(), bson.M{
		"product_id": productID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var reviews []models.Review
	for cursor.Next(context.Background()) {
		var review models.Review
		cursor.Decode(&review)
		reviews = append(reviews, review)
	}
	return reviews, nil
}
