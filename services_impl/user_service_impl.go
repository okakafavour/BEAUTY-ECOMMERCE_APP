package servicesimpl

import (
	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type userServiceImpl struct {
	collection *mongo.Collection
}

func NewUserService() services.UserService {
	return &userServiceImpl{
		collection: config.GetCollection("users"),
	}
}

func (s *userServiceImpl) Register(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if email exists
	count, err := s.collection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return errors.New("failed to check existing user")
	}

	if count > 0 {
		return errors.New("email already registered")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = string(hashed)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Insert into DB
	_, err = s.collection.InsertOne(ctx, user)
	if err != nil {
		return errors.New("failed to create user")
	}

	return nil
}

func (s *userServiceImpl) Login(email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var found models.User

	// Check if user exists
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&found)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("invalid email or password")
		}
		return "", errors.New("failed to find user")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(found.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Generate JWT
	token, err := utils.GenerateToken(found.Email)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
