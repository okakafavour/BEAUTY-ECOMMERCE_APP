package services

import (
	"beauty-ecommerce-backend/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	Register(user models.User) error
	Login(email, password string) (string, error)

	GetAllUsers() ([]models.User, error)
	UpdateUser(userID string, update models.User) error
	DeleteUser(userID string) error
	GetUserByID(id primitive.ObjectID) (models.User, error)
	GetProfile(userID primitive.ObjectID) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)

	// 🔐 Password reset
	SavePasswordResetToken(userID primitive.ObjectID, hashedToken string, expiry time.Time) error
	GetUserByResetToken(hashedToken string) (*models.User, error)
	UpdatePassword(userID primitive.ObjectID, hashedPassword string) error
	ClearResetToken(userID primitive.ObjectID) error
}
