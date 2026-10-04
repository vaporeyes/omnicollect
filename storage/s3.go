// ABOUTME: S3MediaStore implements MediaStore using an S3-compatible object store.
// ABOUTME: Uses AWS SDK v2 with configurable endpoint for AWS S3, MinIO, or R2.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3MediaStore stores images in an S3-compatible object store.
type S3MediaStore struct {
	client *s3.Client
	bucket string
	prefix string
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
func (m *S3MediaStore) SaveOriginal(ctx context.Context, filename string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ValidateFilename(filename); err != nil {
		return err
	}
	return m.upload(ctx, m.prefix+"originals/"+filename, data)
}

// SaveThumbnail uploads thumbnail image bytes to S3.
func (m *S3MediaStore) SaveThumbnail(ctx context.Context, filename string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ValidateFilename(filename); err != nil {
		return err
	}
	return m.upload(ctx, m.prefix+"thumbnails/"+filename, data)
}

// ForTenant isolates object keys without sharing mutable request state.
func (m *S3MediaStore) ForTenant(tenant string) (MediaStore, error) {
	if err := ValidateFilename(tenant); err != nil {
		return nil, err
	}
	return &S3MediaStore{client: m.client, bucket: m.bucket, prefix: "tenants/" + tenant + "/"}, nil
}

// CheckOriginal uses HEAD rather than downloading every original on each edit.
func (m *S3MediaStore) CheckOriginal(ctx context.Context, filename string) error {
	if err := ValidateFilename(filename); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	result, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(m.bucket), Key: aws.String(m.prefix + "originals/" + filename)})
	if err != nil {
		return err
	}
	if result.ContentLength == nil || *result.ContentLength <= 0 || *result.ContentLength > 30<<20 {
		return fmt.Errorf("invalid original object size")
	}
	return nil
}

// GetOriginal downloads an original image from S3.
func (m *S3MediaStore) GetOriginal(ctx context.Context, filename string) ([]byte, error) {
	if err := ValidateFilename(filename); err != nil {
		return nil, err
	}
	return m.download(ctx, m.prefix+"originals/"+filename)
}

// GetThumbnail downloads a thumbnail image from S3.
func (m *S3MediaStore) GetThumbnail(ctx context.Context, filename string) ([]byte, error) {
	if err := ValidateFilename(filename); err != nil {
		return nil, err
	}
	return m.download(ctx, m.prefix+"thumbnails/"+filename)
}

func (m *S3MediaStore) upload(ctx context.Context, key string, data []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err := m.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(m.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(http.DetectContentType(data)),
	})
	if err != nil {
		return fmt.Errorf("uploading to S3 (%s): %w", key, err)
	}
	return nil
}

func (m *S3MediaStore) download(ctx context.Context, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	output, err := m.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("downloading from S3 (%s): %w", key, err)
	}
	defer output.Body.Close()

	limit := int64(30 << 20)
	if strings.Contains(key, "thumbnails/") {
		limit = 2 << 20
	}
	data, err := io.ReadAll(io.LimitReader(output.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("reading S3 object (%s): %w", key, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("media object exceeds size limit")
	}
	return data, nil
}
