package repositories

import (
	"beauty-ecommerce-backend/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductRepository struct {
	Collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) *ProductRepository {
	return &ProductRepository{
		Collection: db.Collection("products"),
	}
}

// CREATE
func (r *ProductRepository) Create(product *models.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.Collection.InsertOne(ctx, product)
	if err != nil {
		return err
	}

	product.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FIND ALL
func (r *ProductRepository) FindAll() ([]models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// FIND BY ID
func (r *ProductRepository) FindByID(id primitive.ObjectID) (*models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var product models.Product
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// UPDATE
func (r *ProductRepository) Update(id primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

// === DELETE ===//
func (r *ProductRepository) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("product not found")
	}
	return nil
}
