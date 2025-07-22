package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	MaxFileSize          = 10 * 1024 * 1024 // 10MB
	MaxSupportingDocSize = 5 * 1024 * 1024  // 5MB for supporting docs
	S3Bucket             = "tayaria-warranty"
	S3BaseURL            = "https://tayaria-warranty.s3.amazonaws.com/"
	FallbackReceipt      = "https://placeholder.com/receipt.pdf"
)

// UploadReceiptToS3 handles file validation, conversion, and upload
func UploadReceiptToS3(file multipart.File, header *multipart.FileHeader, carPlate string) (string, error) {
	// Check file size
	if header.Size > MaxFileSize {
		return FallbackReceipt, fmt.Errorf("file too large (max 10MB)")
	}

	// Read file into buffer
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return FallbackReceipt, fmt.Errorf("failed to read file: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	var contentType string
	var uploadBuf *bytes.Buffer
	var finalExt string

	switch ext {
	case ".pdf":
		contentType = "application/pdf"
		uploadBuf = buf
		finalExt = "pdf"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
		uploadBuf = buf
		finalExt = "jpg"
	case ".png":
		// Convert PNG to JPG
		img, err := png.Decode(bytes.NewReader(buf.Bytes()))
		if err != nil {
			return FallbackReceipt, fmt.Errorf("failed to decode PNG: %v", err)
		}
		jpgBuf := new(bytes.Buffer)
		if err := jpeg.Encode(jpgBuf, img, &jpeg.Options{Quality: 85}); err != nil {
			return FallbackReceipt, fmt.Errorf("failed to convert PNG to JPG: %v", err)
		}
		contentType = "image/jpeg"
		uploadBuf = jpgBuf
		finalExt = "jpg"
	default:
		return FallbackReceipt, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Generate file name: receipts/{car_plate}/{DDMMYYYY}_{random4}.{extension}
	now := time.Now()
	dateStr := now.Format("02012006") // DDMMYYYY
	randomID := randomString(4)
	key := fmt.Sprintf("receipts/%s/%s_%s.%s", carPlate, dateStr, randomID, finalExt)

	// Upload to S3
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return FallbackReceipt, fmt.Errorf("failed to load AWS config: %v", err)
	}
	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(S3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(uploadBuf.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return FallbackReceipt, fmt.Errorf("failed to upload to S3: %v", err)
	}

	return S3BaseURL + key, nil
}

// UploadSupportingDocToS3 uploads a supporting document to S3
func UploadSupportingDocToS3(file multipart.File, header *multipart.FileHeader, claimID string) (string, error) {
	if header.Size > MaxSupportingDocSize {
		return "", fmt.Errorf("file too large (max 5MB)")
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" {
		return "", fmt.Errorf("only PDF files are supported for supporting documents")
	}

	now := time.Now()
	dateStr := now.Format("02012006") // DDMMYYYY
	randomID := randomString(4)
	key := fmt.Sprintf("supporting_docs/%s/%s_%s.pdf", claimID, dateStr, randomID)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %v", err)
	}
	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(S3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}
	return S3BaseURL + key, nil
}

// GeneratePresignedURL generates a pre-signed URL for a given S3 key, valid for the specified duration
func GeneratePresignedURL(s3Key string, expiresIn time.Duration) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	presignInput := &s3.GetObjectInput{
		Bucket: aws.String(S3Bucket),
		Key:    aws.String(s3Key),
	}
	presignResult, err := presigner.PresignGetObject(context.TODO(), presignInput, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", err
	}
	return presignResult.URL, nil
}

// randomString generates a random alphanumeric string of length n
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := 0; i < n; i++ {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}
