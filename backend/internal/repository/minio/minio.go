package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"revisitr/internal/usecase/storage"
)

type Client struct {
	client   *minio.Client
	endpoint string
}

func New(endpoint, accessKey, secretKey string, useSSL bool) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}
	return &Client{client: mc, endpoint: endpoint}, nil
}

func (c *Client) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error) {
	exists, err := c.client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("minio.Upload check bucket: %w", err)
	}
	if !exists {
		if err := c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio.Upload create bucket: %w", err)
		}
	}

	_, err = c.client.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.Upload: %w", err)
	}

	return &storage.FileInfo{
		Key:         key,
		URL:         fmt.Sprintf("/revisitr/storage/%s", key),
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (c *Client) Delete(ctx context.Context, bucket, key string) error {
	return c.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (c *Client) GetURL(_ context.Context, bucket, key string) (string, error) {
	return fmt.Sprintf("/revisitr/storage/%s", key), nil
}
