package storage

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestGeneratePresignedURLs(t *testing.T) {
	client := &Client{
		s3: s3.New(s3.Options{
			BaseEndpoint: aws.String("http://s3.local.test"),
			Region:       "us-east-1",
			Credentials:  credentials.NewStaticCredentialsProvider("access", "secret", ""),
			UsePathStyle: true,
		}),
		bucket: "zoora",
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	uploadURL, err := client.GeneratePresignedUploadURL(context.Background(), "classes/1/file.pdf", 15*time.Minute)
	if err != nil {
		t.Fatalf("GeneratePresignedUploadURL() error = %v", err)
	}
	assertPresignedURL(t, uploadURL, "PUT", "classes/1/file.pdf")

	downloadURL, err := client.GeneratePresignedDownloadURL(context.Background(), "classes/1/file.pdf", time.Hour)
	if err != nil {
		t.Fatalf("GeneratePresignedDownloadURL() error = %v", err)
	}
	assertPresignedURL(t, downloadURL, "GET", "classes/1/file.pdf")
}

func assertPresignedURL(t *testing.T, rawURL, method, key string) {
	t.Helper()
	for _, want := range []string{
		"http://s3.local.test/zoora/" + key,
		"X-Amz-Algorithm=AWS4-HMAC-SHA256",
		"X-Amz-Credential=",
		"X-Amz-SignedHeaders=host",
	} {
		if !strings.Contains(rawURL, want) {
			t.Fatalf("presigned %s URL %q missing %q", method, rawURL, want)
		}
	}
}
