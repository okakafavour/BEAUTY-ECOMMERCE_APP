package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"strings"
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
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}

// UploadRemoteImageWithID uploads an image from a remote URL and returns both the URL and public ID
func UploadRemoteImageWithID(remoteURL string) (string, string, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	// Generate unique public ID
	publicID := fmt.Sprintf("products_%d", time.Now().UnixNano())

	uploadResult, err := cld.Upload.Upload(
		context.Background(),
		remoteURL,
		uploader.UploadParams{
			Folder:   "products",
			PublicID: publicID,
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload remote image: %v", err)
	}

	fmt.Println("✅ Uploaded image to Cloudinary:", uploadResult.SecureURL)
	return uploadResult.SecureURL, publicID, nil
}

// UploadToCloudinaryWithID uploads a multipart file and returns URL + public ID
func UploadToCloudinaryWithID(fileHeader *multipart.FileHeader) (string, string, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return "", "", fmt.Errorf("failed to init Cloudinary: %v", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	uploadResult, err := cld.Upload.Upload(
		context.Background(),
		file,
		uploader.UploadParams{
			Folder:   "products",
			PublicID: fmt.Sprintf("products_%d", time.Now().UnixNano()),
		},
	)
	if err != nil {
		return "", "", err
	}

	return uploadResult.SecureURL, uploadResult.PublicID, nil
}

// DeleteImageFromCloudinaryByURL deletes an image using its full Cloudinary URL
func DeleteImageFromCloudinaryByURL(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("image URL is empty")
	}

	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid image URL: %v", err)
	}

	// Example Cloudinary URL:
	// https://res.cloudinary.com/<cloud>/image/upload/v123456/products/products_12345.png

	parts := strings.Split(parsedURL.Path, "/")

	// Remove leading empty element caused by starting "/"
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	// Expected format:
	// [res, cloudinary, com, <cloud>, image, upload, v123456, products, products_12345.png]
	if len(parts) < 7 {
		return fmt.Errorf("invalid Cloudinary path format")
	}

	// Remove version (v12345)
	filtered := []string{}
	for _, part := range parts {
		if !strings.HasPrefix(part, "v") {
			filtered = append(filtered, part)
		}
	}

	// Public ID = everything after "upload/"
	index := 0
	for i, p := range filtered {
		if p == "upload" {
			index = i + 1
			break
		}
	}

	publicIDWithExt := strings.Join(filtered[index:], "/")
	publicID := strings.TrimSuffix(publicIDWithExt, path.Ext(publicIDWithExt))

	// Initialize Cloudinary with CORRECT VALUES
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize cloudinary: %v", err)
	}

	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		return fmt.Errorf("cloudinary delete failed: %v", err)
	}

	fmt.Println("✅ Deleted image:", publicID)
	return nil
}
