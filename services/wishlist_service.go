package services

import (
	"beauty-ecommerce-backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishlistService interface {
	GetWishlist(userID primitive.ObjectID) (*models.Wishlist, error)
	AddProduct(userID, productID primitive.ObjectID) error
	RemoveProduct(userID, productID primitive.ObjectID) error
	GetWishlistPaginated(userID primitive.ObjectID, page, limit int) ([]models.Product, int64, error)
}
