package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/4H1R/zoora/internal/config"
)

type Client struct {
	s3     *s3.Client
	bucket string
	logger *slog.Logger
}

func NewClient(cfg *config.Config, logger *slog.Logger) (*Client, error) {
	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.S3Endpoint),
		Region:       cfg.S3Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		UsePathStyle: true,
	})

	// Ensure bucket exists
	_, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(cfg.S3Bucket),
	})
	if err != nil {
		// Try to create the bucket in development
		if cfg.IsDevelopment() {
			_, createErr := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
				Bucket: aws.String(cfg.S3Bucket),
			})
			if createErr != nil {
				return nil, fmt.Errorf("creating S3 bucket: %w", createErr)
			}
			logger.Info("created S3 bucket", "bucket", cfg.S3Bucket)
		} else {
			return nil, fmt.Errorf("S3 bucket not accessible: %w", err)
		}
	}

	logger.Info("S3 storage client initialized", "bucket", cfg.S3Bucket)
	return &Client{s3: s3Client, bucket: cfg.S3Bucket, logger: logger}, nil
}

func (c *Client) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(c.s3)
	req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generating presigned upload URL: %w", err)
	}
	return req.URL, nil
}

func (c *Client) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(c.s3)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generating presigned download URL: %w", err)
	}
	return req.URL, nil
}

func (c *Client) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	resp, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("heading object: %w", err)
	}
	return resp, nil
}

func (c *Client) HeadBucket(ctx context.Context) error {
	_, err := c.s3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	return err
}
