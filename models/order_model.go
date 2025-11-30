package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	Items            []OrderItem        `bson:"items" json:"items"`
	Total            float64            `bson:"total" json:"total"`
	Status           string             `bson:"status" json:"status"`
	PaymentReference string             `bson:"payment_reference" json:"payment_reference"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
	TotalPrice       float64            `bson:"total_price" json:"total_price"`
}

type OrderItem struct {
	ProductID   string  ` json:"product_id"`
	ProductName string  `bson:"product_name" json:"product_name"`
	Quantity    int     `bson:"quantity" json:"quantity"`
	Price       float64 `bson:"price" json:"price"`
}
