package services

import (
	"beauty-ecommerce-backend/models"
	"beauty-ecommerce-backend/repositories"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewService struct {
	repo *repositories.ReviewRepository
}

func NewReviewService(repo *repositories.ReviewRepository) *ReviewService {
	return &ReviewService{repo}
}

// -------------------------------
// Create Review
// -------------------------------
func (s *ReviewService) CreateReview(userID, productID primitive.ObjectID, rating int, body string) error {
	review := models.Review{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		ProductID: productID,
		Rating:    rating,
		Body:      body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.repo.Create(&review)
}

// -------------------------------
// Update Review (owner or admin)
// -------------------------------
func (s *ReviewService) UpdateReview(reviewID primitive.ObjectID, userID primitive.ObjectID, isAdmin bool, rating int, body string) error {

	review, err := s.repo.FindByID(reviewID)
	if err != nil {
		return errors.New("review not found")
	}

	// owner-only or admin
	if review.UserID != userID && !isAdmin {
		return errors.New("unauthorized: you can only update your own review")
	}

	update := bson.M{
		"rating":     rating,
		"body":       body,
		"updated_at": time.Now(),
	}

	return s.repo.Update(reviewID, update)
}

// -------------------------------
// Delete Review
// -------------------------------
func (s *ReviewService) DeleteReview(reviewID primitive.ObjectID, userID primitive.ObjectID, isAdmin bool) error {

	review, err := s.repo.FindByID(reviewID)
	if err != nil {
		return errors.New("review not found")
	}

	if review.UserID != userID && !isAdmin {
		return errors.New("unauthorized: you can only delete your own review")
	}

	return s.repo.Delete(reviewID)
}

// -------------------------------
// Get Reviews for Product
// -------------------------------
func (s *ReviewService) GetProductReviews(productID primitive.ObjectID) ([]models.Review, error) {
	return s.repo.GetByProduct(productID)
}
