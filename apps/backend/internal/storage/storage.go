package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client interface {
	PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, key string) (*minio.Object, error)
	CreateSignedUploadURL(ctx context.Context, key string, expires time.Duration, contentType string) (string, error)
	CreateSignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error)
	StatObject(ctx context.Context, key string) error
	DeleteObject(ctx context.Context, key string) error
}

type MinIOClient struct {
	client *minio.Client
	bucket string
}

type Config struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	UsePathStyle bool
	Bucket       string
}

func New(cfg Config) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       "us-east-1",
		BucketLookup: minio.BucketLookupAuto,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &MinIOClient{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (c *MinIOClient) PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (c *MinIOClient) GetObject(ctx context.Context, key string) (*minio.Object, error) {
	object, err := c.client.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (c *MinIOClient) CreateSignedUploadURL(ctx context.Context, key string, expires time.Duration, contentType string) (string, error) {
	uploadURL, err := c.client.PresignedPutObject(ctx, c.bucket, key, expires)
	if err != nil {
		return "", err
	}
	return uploadURL.String(), nil
}

func (c *MinIOClient) CreateSignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	downloadURL, err := c.client.PresignedGetObject(ctx, c.bucket, key, expires, nil)
	if err != nil {
		return "", err
	}
	return downloadURL.String(), nil
}

func (c *MinIOClient) StatObject(ctx context.Context, key string) error {
	_, err := c.client.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	return err
}

func (c *MinIOClient) DeleteObject(ctx context.Context, key string) error {
	return c.client.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}

func ObjectKey(orgID string, documentID string, versionID string, kind string) string {
	return path.Join("orgs", orgID, "documents", documentID, "versions", versionID, kind)
}

func ParseEndpoint(raw string) (string, bool, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false, err
	}
	return parsed.Host, parsed.Scheme == "https", nil
}
