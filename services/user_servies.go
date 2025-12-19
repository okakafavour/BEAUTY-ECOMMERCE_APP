package services

import (
	"beauty-ecommerce-backend/models"

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
}
