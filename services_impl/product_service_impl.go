package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
	"beauty-ecommerce-backend/utils"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productServiceImpl struct {
	productRepo *repositories.ProductRepository
}

func NewProductService(productRepo *repositories.ProductRepository) services.ProductService {
	return &productServiceImpl{
		productRepo: productRepo,
	}
}

// CREATE PRODUCT
func (s *productServiceImpl) CreateProduct(product *models.Product) error {
	if product.ID.IsZero() {
		product.ID = primitive.NewObjectID()
	}
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	return s.productRepo.Create(product)
}

// GET ALL PRODUCTS
func (s *productServiceImpl) GetAllProducts() ([]models.Product, error) {
	return s.productRepo.FindAll()
}

// GET PRODUCT BY ID
func (s *productServiceImpl) GetProductByID(id string) (*models.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}

	product, err := s.productRepo.FindByID(objID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	return product, nil
}

// UPDATE PRODUCT
func (s *productServiceImpl) UpdateProduct(id string, product models.Product) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID")
	}

	update := bson.M{}

	// Only update fields that are non-empty / non-zero
	if product.Name != "" {
		update["name"] = product.Name
	}
	if product.Description != "" {
		update["description"] = product.Description
	}
	if product.Price != 0 {
		update["price"] = product.Price
	}
	if product.Stock != 0 {
		update["stock"] = product.Stock
	}
	if product.Category != "" {
		update["category"] = product.Category
	}
	if product.ImageURL != "" {
		update["image_url"] = product.ImageURL
	}

	// Always update timestamp
	update["updated_at"] = time.Now()

	return s.productRepo.Update(objID, update)
}

// DELETE PRODUCT
// DELETE PRODUCT
func (s *productServiceImpl) DeleteProduct(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID")
	}

	// Fetch the product first
	product, err := s.productRepo.FindByID(objID)
	if err != nil {
		return errors.New("product not found")
	}

	// Delete image from Cloudinary if exists
	if product.ImageID != "" {
		if err := utils.DeleteImageFromCloudinary(product.ImageID); err != nil {
			fmt.Println("⚠️ failed to delete image from Cloudinary:", err)
			// Don't block deletion, just log
		}
	}

	// Delete product from DB
	return s.productRepo.Delete(objID)
}
