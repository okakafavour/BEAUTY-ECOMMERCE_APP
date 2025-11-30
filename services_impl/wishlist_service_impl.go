package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishlistService struct {
	repo *repositories.WishlistRepository
}

func NewWishlistService(repo *repositories.WishlistRepository) *WishlistService {
	return &WishlistService{repo}
}

// -------------------------------
// Get wishlist for a user
// -------------------------------
func (s *WishlistService) GetWishlist(userID primitive.ObjectID) (*models.Wishlist, error) {
	wishlist, err := s.repo.FindByUser(userID)
	if err != nil {
		// Create empty wishlist if not exists
		wishlist = &models.Wishlist{
			ID:         primitive.NewObjectID(),
			UserID:     userID,
			ProductIDs: []primitive.ObjectID{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err = s.repo.Create(wishlist)
		if err != nil {
			return nil, err
		}
	}
	return wishlist, nil
}

// -------------------------------
// Add product to wishlist
// -------------------------------
func (s *WishlistService) AddProduct(userID, productID primitive.ObjectID) error {
	wishlist, err := s.GetWishlist(userID)
	if err != nil {
		return err
	}

	// Prevent duplicates
	for _, pid := range wishlist.ProductIDs {
		if pid == productID {
			return errors.New("product already in wishlist")
		}
	}

	wishlist.ProductIDs = append(wishlist.ProductIDs, productID)
	wishlist.UpdatedAt = time.Now()

	return s.repo.UpdateProducts(userID, wishlist.ProductIDs)
}

// -------------------------------
// Remove product from wishlist
// -------------------------------
func (s *WishlistService) RemoveProduct(userID, productID primitive.ObjectID) error {
	wishlist, err := s.GetWishlist(userID)
	if err != nil {
		return err
	}

	newProductIDs := []primitive.ObjectID{}
	for _, pid := range wishlist.ProductIDs {
		if pid != productID {
			newProductIDs = append(newProductIDs, pid)
		}
	}

	wishlist.ProductIDs = newProductIDs
	wishlist.UpdatedAt = time.Now()

	return s.repo.UpdateProducts(userID, wishlist.ProductIDs)
}
