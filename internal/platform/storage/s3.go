package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/4H1R/zoora/internal/config"
)

type Client struct {
	s3             *s3.Client
	presign        *s3.Client
	bucket         string
	publicBucket   string
	publicEndpoint string
	logger         *slog.Logger
}

func NewClient(cfg *config.Config, logger *slog.Logger) (*Client, error) {
	opts := s3.Options{
		Region:       cfg.S3Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		UsePathStyle: true,
	}

	internalOpts := opts
	internalOpts.BaseEndpoint = aws.String(cfg.S3Endpoint)
	s3Client := s3.New(internalOpts)

	// Presigned URLs are handed to browsers, so they must carry the public host.
	// The presign client never dials S3 (signing is local), so pointing it at
	// the public endpoint costs no connectivity at boot.
	publicEndpoint := cfg.S3PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.S3Endpoint
	}
	publicOpts := opts
	publicOpts.BaseEndpoint = aws.String(publicEndpoint)
	presignClient := s3.New(publicOpts)

	_, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(cfg.S3Bucket),
	})
	if err != nil {
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

	publicBucket := cfg.S3PublicBucket
	if publicBucket == "" {
		publicBucket = cfg.S3Bucket // fallback: not actually public, dev-only
	}
	if publicBucket != cfg.S3Bucket {
		if _, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
			Bucket: aws.String(publicBucket),
		}); err != nil {
			if cfg.IsDevelopment() {
				if _, cErr := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
					Bucket: aws.String(publicBucket),
				}); cErr != nil {
					return nil, fmt.Errorf("creating public S3 bucket: %w", cErr)
				}
				logger.Info("created public S3 bucket", "bucket", publicBucket)
			} else {
				return nil, fmt.Errorf("public S3 bucket not accessible: %w", err)
			}
		}
		// Best-effort anonymous read policy so permanent URLs resolve in the
		// browser. RustFS/S3 accept a standard bucket policy; a failure here is
		// logged, not fatal (policy may be managed out-of-band in prod).
		policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"s3:GetObject","Resource":"arn:aws:s3:::%s/*"}]}`, publicBucket)
		if _, err := s3Client.PutBucketPolicy(context.Background(), &s3.PutBucketPolicyInput{
			Bucket: aws.String(publicBucket),
			Policy: aws.String(policy),
		}); err != nil {
			logger.Warn("set public bucket policy failed (set manually if assets 403)", "bucket", publicBucket, "error", err)
		}
	}

	logger.Info("S3 storage client initialized", "bucket", cfg.S3Bucket, "public_bucket", publicBucket)
	return &Client{
		s3:             s3Client,
		presign:        presignClient,
		bucket:         cfg.S3Bucket,
		publicBucket:   publicBucket,
		publicEndpoint: publicEndpoint,
		logger:         logger,
	}, nil
}

func (c *Client) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(c.presign)
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
	presigner := s3.NewPresignClient(c.presign)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generating presigned download URL: %w", err)
	}
	return req.URL, nil
}

// GetObject downloads an object's full contents. Intended for small
// server-processed files (e.g. import spreadsheets).
//
// maxSize bounds how much is ever read into memory: a maxSize <= 0 means no
// limit. When positive, a lying/absent Content-Length is not trusted alone —
// the body is also wrapped in an io.LimitReader capped at maxSize+1 so a
// read that hits the extra byte proves the object exceeds the limit even if
// S3 reported a smaller (or chunked/missing) size upfront.
func (c *Client) GetObject(ctx context.Context, key string, maxSize int64) ([]byte, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("storage.GetObject %s: %w", key, err)
	}
	defer out.Body.Close()

	if maxSize > 0 {
		if out.ContentLength != nil && aws.ToInt64(out.ContentLength) > maxSize {
			return nil, fmt.Errorf("storage.GetObject %s: object size %d exceeds limit %d", key, aws.ToInt64(out.ContentLength), maxSize)
		}
		data, err := io.ReadAll(io.LimitReader(out.Body, maxSize+1))
		if err != nil {
			return nil, fmt.Errorf("storage.GetObject read %s: %w", key, err)
		}
		if int64(len(data)) > maxSize {
			return nil, fmt.Errorf("storage.GetObject %s: object exceeds limit %d", key, maxSize)
		}
		return data, nil
	}

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("storage.GetObject read %s: %w", key, err)
	}
	return data, nil
}

// PutObject uploads bytes to the bucket under key with the given content type.
// Used for server-side writes (e.g. generated invoice-receipt PDFs) that never
// pass through a browser presign.
func (c *Client) PutObject(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("putting object %s: %w", key, err)
	}
	return nil
}

// DeleteObject removes an object from the bucket. S3 delete is idempotent —
// deleting a missing key returns success — so callers need not pre-check.
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	return nil
}

// DeleteByPrefix removes every object whose key begins with prefix, paging
// through the listing and batch-deleting up to 1000 keys per call (the S3
// DeleteObjects limit). Used to purge a whole tenant's storage on org delete.
// Idempotent: an empty listing is a no-op.
func (c *Client) DeleteByPrefix(ctx context.Context, prefix string) error {
	paginator := s3.NewListObjectsV2Paginator(c.s3, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing objects for prefix %s: %w", prefix, err)
		}
		if len(page.Contents) == 0 {
			continue
		}
		ids := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			ids = append(ids, types.ObjectIdentifier{Key: obj.Key})
		}
		if _, err := c.s3.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(c.bucket),
			Delete: &types.Delete{Objects: ids, Quiet: aws.Bool(true)},
		}); err != nil {
			return fmt.Errorf("deleting objects for prefix %s: %w", prefix, err)
		}
	}
	return nil
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

// PublicPresignUpload returns a presigned PUT URL that targets the public
// bucket. The uploaded object is world-readable via PublicURL(key).
func (c *Client) PublicPresignUpload(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(c.presign)
	req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.publicBucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generating public presigned upload URL: %w", err)
	}
	return req.URL, nil
}

// PublicURL builds the permanent, non-signed browser URL for a public object.
// Uses the public endpoint (never the internal S3 host browsers can't reach).
func (c *Client) PublicURL(key string) string {
	return strings.TrimRight(c.publicEndpoint, "/") + "/" + c.publicBucket + "/" + key
}

// DeletePublicObject removes an object from the public bucket. Idempotent.
func (c *Client) DeletePublicObject(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.publicBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting public object: %w", err)
	}
	return nil
}
