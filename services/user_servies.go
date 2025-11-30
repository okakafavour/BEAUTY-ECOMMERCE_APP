package services

import "beauty-ecommerce-backend/models"

type UserService interface {
	Register(user models.User) error
	Login(email, password string) (string, error)

	GetAllUsers() ([]models.User, error)
	UpdateUser(userID string, update models.User) error
	DeleteUser(userID string) error
}
