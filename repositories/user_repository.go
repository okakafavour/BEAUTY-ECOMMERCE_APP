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

// Find user by ID (updated to use ObjectID)
func (r *UserRepository) FindById(id string) (models.User, error) {
	var user models.User

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return user, errors.New("invalid user id")
	}

	err = r.Collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user)
	return user, err
}

// ADMIN: Find all users
func (r *UserRepository) FindAll() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// ADMIN: Update user
func (r *UserRepository) Update(userID string, update models.User) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	updateBson := bson.M{
		"$set": bson.M{
			"name":       update.Name,
			"email":      update.Email,
			"role":       update.Role,
			"updated_at": time.Now(),
		},
	}

	res, err := r.Collection.UpdateOne(context.Background(), bson.M{"_id": id}, updateBson)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ADMIN: Delete user
func (r *UserRepository) Delete(userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	res, err := r.Collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) FindByID(userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.Collection.FindOne(
		context.TODO(),
		bson.M{"_id": userID},
	).Decode(&user)

	if err != nil {
		return nil, err
	}
	return &user, nil
}
