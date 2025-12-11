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

func BoolPtr(b bool) *bool {
	return &b
}

func makeTimestamp() int64 {
	return int64(float64(time.Now().UnixNano()) / 1e6)
}

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

func UploadRemoteImage(remoteURL string) (string, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	publicID := fmt.Sprintf("image_%d", time.Now().UnixNano())

	uploadResult, err := cld.Upload.Upload(
		context.Background(),
		remoteURL,
		uploader.UploadParams{
			Folder:   "products",
			PublicID: publicID,
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

func UploadRemoteImageWithID(remoteURL string) (string, string, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

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

	finalURL := uploadResult.SecureURL
	finalID := uploadResult.PublicID

	if finalID == "" && finalURL != "" {
		parts := strings.Split(finalURL, "/")

		filename := parts[len(parts)-1]

		uploadIndex := -1
		for i, p := range parts {
			if p == "upload" {
				uploadIndex = i
				break
			}
		}

		if uploadIndex != -1 && uploadIndex+2 < len(parts)-1 {
			folderParts := parts[uploadIndex+2 : len(parts)-1]
			folder := strings.Join(folderParts, "/")

			finalID = folder + "/" + strings.TrimSuffix(filename, path.Ext(filename))
		}
	}

	return finalURL, finalID, nil
}

func DeleteImageFromCloudinaryByURL(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("image URL is empty")
	}

	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid image URL: %v", err)
	}

	parts := strings.Split(parsedURL.Path, "/")

	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	if len(parts) < 7 {
		return fmt.Errorf("invalid Cloudinary path format")
	}

	filtered := []string{}
	for _, part := range parts {
		if !strings.HasPrefix(part, "v") {
			filtered = append(filtered, part)
		}
	}

	index := 0
	for i, p := range filtered {
		if p == "upload" {
			index = i + 1
			break
		}
	}

	publicIDWithExt := strings.Join(filtered[index:], "/")
	publicID := strings.TrimSuffix(publicIDWithExt, path.Ext(publicIDWithExt))

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
