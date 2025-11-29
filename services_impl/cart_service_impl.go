package servicesimpl

import (
	"context"
	"errors"
	"time"

	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartServiceImpl struct {
	cartRepo *repositories.CartRepository
}

func NewCartService(cartRepo *repositories.CartRepository) *CartServiceImpl {
	return &CartServiceImpl{cartRepo: cartRepo}
}

func (c *CartServiceImpl) CreateCartItem(cartItem models.CartItem) (models.CartItem, error) {
	if cartItem.Quantity <= 0 {
		return models.CartItem{}, errors.New("quantity must be greater than zero")
	}

	cartItem.ID = primitive.NewObjectID()
	cartItem.CreatedAt = time.Now()
	cartItem.UpdatedAt = time.Now()

	err := c.cartRepo.AddToCart(context.Background(), &cartItem)
	return cartItem, err
}

func (c *CartServiceImpl) GetCartByUser(userID primitive.ObjectID) ([]models.CartItem, error) {
	return c.cartRepo.GetUserCart(context.Background(), userID)
}

func (c *CartServiceImpl) UpdateCartItem(cartItem models.CartItem) (models.CartItem, error) {
	cartItem.UpdatedAt = time.Now()
	err := c.cartRepo.UpdateQuantity(context.Background(), cartItem.ID, cartItem.Quantity)
	return cartItem, err
}

func (c *CartServiceImpl) DeleteCartItem(cartItemID primitive.ObjectID) error {
	return c.cartRepo.DeleteCartItem(context.Background(), cartItemID)
}

func (c *CartServiceImpl) ClearCart(userID primitive.ObjectID) error {
	cartItems, err := c.cartRepo.GetUserCart(context.Background(), userID)
	if err != nil {
		return err
	}
	for _, item := range cartItems {
		if err := c.cartRepo.DeleteCartItem(context.Background(), item.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *CartServiceImpl) GetCartItemByID(cartItemID primitive.ObjectID) (*models.CartItem, error) {
	return s.cartRepo.FindByID(cartItemID)
}
