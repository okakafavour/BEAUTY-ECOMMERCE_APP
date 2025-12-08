package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	CustomerName     string             `bson:"customer_name" json:"customer_name"`
	CustomerEmail    string             `bson:"customer_email" json:"customer_email"`
	CustomerPhone    string             `bson:"customer_phone" json:"customer_phone"`
	ShippingAddress  Address            `bson:"shipping_address" json:"shipping_address"`
	Items            []OrderItem        `bson:"items" json:"items"`
	Subtotal         float64            `bson:"subtotal" json:"subtotal"`
	ShippingFee      float64            `bson:"shipping_fee" json:"shipping_fee"`
	TotalPrice       float64            `bson:"total_price" json:"total_price"`
	Status           string             `bson:"status" json:"status"`
	PaymentReference string             `bson:"payment_reference" json:"payment_reference"`
	PaymentStatus    string             `bson:"payment_status" json:"payment_status"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type Address struct {
	Street     string `bson:"street" json:"street"`
	City       string `bson:"city" json:"city"`
	State      string `bson:"state" json:"state"`
	Country    string `bson:"country" json:"country"`
	PostalCode string `bson:"postal_code" json:"postal_code"`
}

type OrderItem struct {
	ProductID   string  `bson:"product_id" json:"product_id"`
	ProductName string  `bson:"product_name" json:"product_name"`
	Quantity    int     `bson:"quantity" json:"quantity"`
	Price       float64 `bson:"price" json:"price"`
}
