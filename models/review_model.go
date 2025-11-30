package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ProductID primitive.ObjectID `bson:"product_id" json:"product_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Rating    int                `bson:"rating" json:"rating"` // 1-5
	Title     string             `bson:"title,omitempty" json:"title,omitempty"`
	Body      string             `bson:"body,omitempty" json:"body,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
