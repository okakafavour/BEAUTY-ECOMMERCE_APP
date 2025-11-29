package services

import (
	"beauty-ecommerce-backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartService interface {
	CreateCartItem(cartItem models.CartItem) (models.CartItem, error)
	GetCartByUser(userID primitive.ObjectID) ([]models.CartItem, error)
	UpdateCartItem(cartItem models.CartItem) (models.CartItem, error)
	DeleteCartItem(cartItemID primitive.ObjectID) error
	ClearCart(userID primitive.ObjectID) error
	GetCartItemByID(cartItemID primitive.ObjectID) (*models.CartItem, error) // needed for controller
}
