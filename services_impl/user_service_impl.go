package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type userServiceImpl struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) services.UserService {
	return &userServiceImpl{
		userRepo: userRepo,
	}
}

// -------------------- REGISTER --------------------
func (s *userServiceImpl) Register(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if user.Email == "" || user.Password == "" {
		return errors.New("email and password are required")
	}

	count, err := s.userRepo.Collection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return errors.New("failed to check existing user")
	}
	if count > 0 {
		return errors.New("email already registered")
	}

	if user.Role == "" {
		user.Role = "USER"
	} else {
		user.Role = strings.ToUpper(user.Role)
	}

	if user.Role != "ADMIN" && user.Role != "USER" {
		return errors.New("invalid role (allowed: ADMIN, USER)")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = string(hashed)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Insert user and get inserted ID
	res, err := s.userRepo.Collection.InsertOne(ctx, user)
	if err != nil {
		return errors.New("failed to create user")
	}

	// Ensure inserted ID is ObjectID
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if ok {
		user.ID = oid
	}

	return nil
}

// -------------------- LOGIN --------------------
func (s *userServiceImpl) Login(email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var found models.User
	err := s.userRepo.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&found)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("invalid email or password")
		}
		return "", errors.New("failed to find user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(found.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Generate JWT with correct MongoDB user ID
	token, err := utils.GenerateToken(found.ID, found.Email, found.Role)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

// -------------------- ADMIN METHODS --------------------
func (s *userServiceImpl) GetAllUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.userRepo.Collection.Find(ctx, bson.M{})
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

func (s *userServiceImpl) UpdateUser(userID string, update models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	updateBson := bson.M{
		"$set": bson.M{
			"name":       update.Name,
			"email":      update.Email,
			"role":       strings.ToUpper(update.Role),
			"updated_at": time.Now(),
		},
	}

	res, err := s.userRepo.Collection.UpdateOne(ctx, bson.M{"_id": id}, updateBson)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *userServiceImpl) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	res, err := s.userRepo.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *userServiceImpl) GetUserByID(id primitive.ObjectID) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := s.userRepo.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

func (s *userServiceImpl) GetProfile(userID primitive.ObjectID) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}
