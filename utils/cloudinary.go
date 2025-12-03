package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// -------------------- HELPERS --------------------

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// makeTimestamp returns current Unix timestamp in milliseconds
func makeTimestamp() int64 {
	return int64(float64(time.Now().UnixNano()) / 1e6)
}

// -------------------- UPLOAD FUNCTIONS --------------------

// UploadFile uploads a local file to Cloudinary and returns the secure URL
func UploadFile(filePath string) (string, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	uploadResult, err := cld.Upload.Upload(
		context.Background(),
		filePath,
		uploader.UploadParams{
			Folder:    "products",
			PublicID:  fmt.Sprintf("products/%d", makeTimestamp()),
			Overwrite: BoolPtr(true),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload local file: %v", err)
	}

	return uploadResult.SecureURL, nil
}

// UploadToCloudinary uploads a file from multipart (form-data) to Cloudinary
func UploadToCloudinary(fileHeader *multipart.FileHeader) (string, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	uploadResult, err := cld.Upload.Upload(
		context.Background(),
		file,
		uploader.UploadParams{
			Folder:    "products",
			PublicID:  fmt.Sprintf("products/%d", makeTimestamp()),
			Overwrite: BoolPtr(true),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	return uploadResult.SecureURL, nil
}

// UploadRemoteImage uploads an image from a remote URL to Cloudinary
func UploadRemoteImage(remoteURL string) (string, error) {
	// Initialize Cloudinary from CLOUDINARY_URL
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	// PublicID for the image (unique)
	publicID := fmt.Sprintf("image_%d", time.Now().UnixNano())

	// Upload to folder "products"
	uploadResult, err := cld.Upload.Upload(
		context.Background(),
		remoteURL,
		uploader.UploadParams{
			Folder:   "products", // just one folder
			PublicID: publicID,   // unique name
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload remote image: %v", err)
	}

	fmt.Println("✅ Uploaded image to Cloudinary:", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}

var ctx = context.Background()

func DeleteImageFromCloudinary(publicID string) error {
	cld, err := cloudinary.NewFromParams(
		"YOUR_CLOUD_NAME",
		"YOUR_API_KEY",
		"YOUR_API_SECRET",
	)
	if err != nil {
		return fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}
