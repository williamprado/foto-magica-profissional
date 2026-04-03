package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type UploadResult struct {
	Key       string `json:"key"`
	PublicURL string `json:"publicUrl"`
}

type Provider interface {
	Upload(ctx context.Context, key string, contentType string, body []byte) (UploadResult, error)
	SignedGetURL(ctx context.Context, key string, expires time.Duration) (string, error)
	Read(ctx context.Context, key string) ([]byte, error)
}

type LocalProvider struct {
	basePath string
}

func NewLocalProvider(basePath string) (Provider, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, err
	}
	return LocalProvider{basePath: basePath}, nil
}

func (p LocalProvider) Upload(_ context.Context, key string, _ string, body []byte) (UploadResult, error) {
	path := filepath.Join(p.basePath, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return UploadResult{}, err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return UploadResult{}, err
	}
	return UploadResult{Key: key, PublicURL: "/storage/" + url.PathEscape(key)}, nil
}

func (p LocalProvider) SignedGetURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "/storage/" + url.PathEscape(key), nil
}

func (p LocalProvider) Read(_ context.Context, key string) ([]byte, error) {
	return os.ReadFile(filepath.Join(p.basePath, key))
}

type S3Provider struct {
	bucket    string
	client    *s3.Client
	presigner *s3.PresignClient
}

func NewS3Provider(ctx context.Context, bucket, region, endpoint, accessKey, secretKey string, usePathStyle bool) (Provider, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
		if endpoint != "" {
			o.BaseEndpoint = &endpoint
		}
	})

	return S3Provider{
		bucket:    bucket,
		client:    client,
		presigner: s3.NewPresignClient(client),
	}, nil
}

func (p S3Provider) Upload(ctx context.Context, key string, contentType string, body []byte) (UploadResult, error) {
	uploader := manager.NewUploader(p.client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      &p.bucket,
		Key:         &key,
		Body:        bytes.NewReader(body),
		ContentType: &contentType,
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return UploadResult{}, err
	}
	url, err := p.SignedGetURL(ctx, key, 15*time.Minute)
	if err != nil {
		return UploadResult{}, err
	}
	return UploadResult{Key: key, PublicURL: url}, nil
}

func (p S3Provider) SignedGetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	result, err := p.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (p S3Provider) Read(ctx context.Context, key string) ([]byte, error) {
	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

func ReadAll(file io.Reader) ([]byte, error) {
	payload, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return payload, nil
}
