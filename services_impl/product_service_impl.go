package servicesimpl

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"beauty-ecommerce-backend/services"
	"errors"
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

	update := bson.M{
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"image_url":   product.ImageURL,
		"category":    product.Category,
		"stock":       product.Stock,
		"updated_at":  time.Now(),
	}

	return s.productRepo.Update(objID, update)
}

// DELETE PRODUCT
func (s *productServiceImpl) DeleteProduct(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID")
	}

	return s.productRepo.Delete(objID)
}
