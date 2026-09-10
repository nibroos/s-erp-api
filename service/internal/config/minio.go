package config

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorage wraps object storage for chat attachments. Two clients are held:
// one bound to the internal endpoint (used for uploads from the service) and
// one bound to the public endpoint (used only to *presign* GET URLs, which is
// an offline operation, so those URLs resolve from the user's browser).
type MinioStorage struct {
	client        *minio.Client
	presignClient *minio.Client
	bucket        string
	urlTTL        time.Duration
}

// NewMinioStorage initialises the clients and ensures the bucket exists.
func NewMinioStorage() (*MinioStorage, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT not set")
	}
	publicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	region := os.Getenv("MINIO_REGION")
	if region == "" {
		region = "us-east-1"
	}
	// The public endpoint may be TLS-terminated (e.g. by nginx in prod) while
	// the internal endpoint is plain HTTP. Default to the internal setting.
	publicUseSSL := useSSL
	if v := os.Getenv("MINIO_PUBLIC_USE_SSL"); v != "" {
		publicUseSSL = v == "true"
	}

	creds := credentials.NewStaticV4(accessKey, secretKey, "")
	client, err := minio.New(endpoint, &minio.Options{Creds: creds, Secure: useSSL, Region: region})
	if err != nil {
		return nil, err
	}
	// The presign client points at the browser-facing endpoint. Setting the
	// region explicitly keeps PresignedGetObject fully offline (otherwise it
	// makes a GetBucketLocation call, which this client cannot reach).
	presignClient, err := minio.New(publicEndpoint, &minio.Options{Creds: creds, Secure: publicUseSSL, Region: region})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinioStorage{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		urlTTL:        24 * time.Hour,
	}, nil
}

// Upload streams an object into the bucket.
func (m *MinioStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Download reads an object into memory, refusing anything larger than maxBytes.
// Used to feed chat image attachments to the vision model, which needs the
// bytes themselves rather than a presigned URL — the model runs in its own
// container and cannot reach this one.
func (m *MinioStorage) Download(ctx context.Context, objectKey string, maxBytes int64) ([]byte, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	// Stat first so an oversized object is rejected before it is read.
	info, err := obj.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size > maxBytes {
		return nil, fmt.Errorf("object %s is %d bytes, over the %d limit", objectKey, info.Size, maxBytes)
	}
	return io.ReadAll(io.LimitReader(obj, maxBytes))
}

// Remove deletes an object (best-effort; missing objects are not an error).
func (m *MinioStorage) Remove(ctx context.Context, objectKey string) error {
	return m.client.RemoveObject(ctx, m.bucket, objectKey, minio.RemoveObjectOptions{})
}

// PresignedURL returns a time-limited GET URL for an object. When download is
// true the response is forced as an attachment with the given file name.
func (m *MinioStorage) PresignedURL(ctx context.Context, objectKey, fileName string, download bool) (string, error) {
	reqParams := url.Values{}
	if download && fileName != "" {
		reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	}
	u, err := m.presignClient.PresignedGetObject(ctx, m.bucket, objectKey, m.urlTTL, reqParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
