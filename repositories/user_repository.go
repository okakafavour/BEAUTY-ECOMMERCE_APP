package repositories

import (
	"beauty-ecommerce-backend/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	Collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		Collection: db.Collection("users"),
	}
}

// Create a new user
func (r *UserRepository) CreateUser(user models.User) (*mongo.InsertOneResult, error) {
	return r.Collection.InsertOne(context.TODO(), user)
}

// Find user by email
func (r *UserRepository) FindByEmail(email string) (models.User, error) {
	var user models.User
	err := r.Collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	return user, err
}

// Find user by ID
func (r *UserRepository) FindById(id string) (models.User, error) {
	var user models.User
	err := r.Collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	return user, err
}
