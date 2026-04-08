// ABOUTME: S3MediaStore implements MediaStore using an S3-compatible object store.
// ABOUTME: Uses AWS SDK v2 with configurable endpoint for AWS S3, MinIO, or R2.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3MediaStore stores images in an S3-compatible object store.
type S3MediaStore struct {
	client *s3.Client
	bucket string
}

// NewS3MediaStore creates an S3MediaStore connected to the given endpoint.
func NewS3MediaStore(endpoint, bucket, accessKey, secretKey, region string) (*S3MediaStore, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Required for MinIO and most S3-compatible services
	})

	return &S3MediaStore{client: client, bucket: bucket}, nil
}

// Ping checks S3 connectivity by listing buckets (lightweight operation).
func (m *S3MediaStore) Ping(ctx context.Context) error {
	_, err := m.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(m.bucket),
	})
	return err
}

// SaveOriginal uploads original image bytes to S3.
func (m *S3MediaStore) SaveOriginal(filename string, data []byte) error {
	return m.upload("originals/"+filename, data)
}

// SaveThumbnail uploads thumbnail image bytes to S3.
func (m *S3MediaStore) SaveThumbnail(filename string, data []byte) error {
	return m.upload("thumbnails/"+filename, data)
}

// OriginalURL returns the backend proxy URL for an original image.
func (m *S3MediaStore) OriginalURL(filename string) string {
	return "/originals/" + filename
}

// ThumbnailURL returns the backend proxy URL for a thumbnail image.
func (m *S3MediaStore) ThumbnailURL(filename string) string {
	return "/thumbnails/" + filename
}

// GetOriginal downloads an original image from S3.
func (m *S3MediaStore) GetOriginal(ctx context.Context, filename string) ([]byte, error) {
	return m.download(ctx, "originals/"+filename)
}

// GetThumbnail downloads a thumbnail image from S3.
func (m *S3MediaStore) GetThumbnail(ctx context.Context, filename string) ([]byte, error) {
	return m.download(ctx, "thumbnails/"+filename)
}

func (m *S3MediaStore) upload(key string, data []byte) error {
	_, err := m.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("uploading to S3 (%s): %w", key, err)
	}
	return nil
}

func (m *S3MediaStore) download(ctx context.Context, key string) ([]byte, error) {
	output, err := m.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("downloading from S3 (%s): %w", key, err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("reading S3 object (%s): %w", key, err)
	}
	return data, nil
}
