package servicesimpl

import (
	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"context"
	"errors"
	"strings"
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

// -------------------- REGISTER --------------------
func (s *userServiceImpl) Register(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate basic fields
	if user.Email == "" || user.Password == "" {
		return errors.New("email and password are required")
	}

	// Check if email already exists
	count, err := s.collection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return errors.New("failed to check existing user")
	}

	if count > 0 {
		return errors.New("email already registered")
	}

	// Set default role if not provided
	if user.Role == "" {
		user.Role = "USER"
	} else {
		user.Role = strings.ToUpper(user.Role)
	}

	// Accept only ADMIN or USER
	if user.Role != "ADMIN" && user.Role != "USER" {
		return errors.New("invalid role (allowed: ADMIN, USER)")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = string(hashed)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Insert user
	_, err = s.collection.InsertOne(ctx, user)
	if err != nil {
		return errors.New("failed to create user")
	}

	return nil
}

// -------------------- LOGIN --------------------
func (s *userServiceImpl) Login(email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var found models.User

	// Find user by email
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&found)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("invalid email or password")
		}
		return "", errors.New("failed to find user")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(found.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Generate JWT with email + role
	token, err := utils.GenerateToken(found.ID, found.Email, found.Role)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
