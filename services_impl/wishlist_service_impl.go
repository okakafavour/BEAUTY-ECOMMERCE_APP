package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishlistServiceImpl struct {
	repo *repositories.WishlistRepository
}

func NewWishlistService(repo *repositories.WishlistRepository) services.WishlistService {
	return &WishlistServiceImpl{repo}
}

func (s *WishlistServiceImpl) GetWishlist(userID primitive.ObjectID) (*models.Wishlist, error) {
	wishlist, err := s.repo.FindByUser(userID)
	if err != nil {
		return &models.Wishlist{
			UserID:     userID,
			ProductIDs: []primitive.ObjectID{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}, nil
	}
	return wishlist, nil
}

func (s *WishlistServiceImpl) AddProduct(userID, productID primitive.ObjectID) error {
	return s.repo.AddProduct(userID, productID)
}

func (s *WishlistServiceImpl) RemoveProduct(userID, productID primitive.ObjectID) error {
	return s.repo.RemoveProduct(userID, productID)
}

func (s *WishlistServiceImpl) GetWishlistPaginated(
	userID primitive.ObjectID,
	page, limit int,
) ([]primitive.ObjectID, int64, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	return s.repo.GetPaginated(userID, offset, limit)
}
