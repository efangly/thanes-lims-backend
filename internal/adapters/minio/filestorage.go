package minio

import (
	"context"
	"io"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Adapter struct {
	client *minio.Client
	bucket string
}

func New(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Adapter, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Adapter{client: client, bucket: bucket}, nil
}

// EnsureBucket creates the configured bucket if it doesn't already exist.
// Call once at API boot.
func (a *Adapter) EnsureBucket(ctx context.Context) error {
	exists, err := a.client.BucketExists(ctx, a.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return a.client.MakeBucket(ctx, a.bucket, minio.MakeBucketOptions{})
	}
	return nil
}

func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := a.client.PutObject(ctx, a.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return a.client.GetObject(ctx, a.bucket, key, minio.GetObjectOptions{})
}

func (a *Adapter) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := a.client.PresignedGetObject(ctx, a.bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (a *Adapter) Delete(ctx context.Context, key string) error {
	return a.client.RemoveObject(ctx, a.bucket, key, minio.RemoveObjectOptions{})
}
