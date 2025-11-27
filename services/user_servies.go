package services

import "beauty-ecommerce-backend/models"

type UserService interface {
	Register(user models.User) error
	Login(email, password string) (string, error)
}
