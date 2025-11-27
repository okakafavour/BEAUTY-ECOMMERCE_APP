package services

import "beauty-ecommerce-backend/models"

type ProductService interface {
	CreateProduct(product *models.Product) error // <-- pointer
	GetAllProducts() ([]models.Product, error)
	GetProductByID(id string) (*models.Product, error)
	UpdateProduct(id string, product models.Product) error
	DeleteProduct(id string) error
}
