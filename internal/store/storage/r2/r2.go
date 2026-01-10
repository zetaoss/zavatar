// internal/store/storage/r2/r2.go
package r2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/zetaoss/zavatar/internal/store/storage"
)

type Config struct {
	AccountID string
	Bucket    string
	AccessKey string
	SecretKey string
	Prefix    string
}

type Storage struct {
	s3     *s3.Client
	bucket string
	prefix string
}

func New(ctx context.Context, cfg Config) (*Storage, error) {
	if cfg.AccountID == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("r2: missing required config")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("r2: load aws config: %w", err)
	}

	cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &Storage{
		s3:     cli,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
	}, nil
}

func (s *Storage) withPrefix(key string) string {
	if s.prefix == "" {
		return key
	}
	return s.prefix + key
}

func (s *Storage) Get(ctx context.Context, key string) ([]byte, error) {
	k := s.withPrefix(key)

	out, err := s.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(k),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	defer func() { _ = out.Body.Close() }()

	b, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Storage) Put(ctx context.Context, key string, contentType string, body []byte) error {
	k := s.withPrefix(key)

	_, err := s.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(k),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	k := s.withPrefix(key)

	_, err := s.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(k),
	})
	if err != nil {
		if isNotFound(err) {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}

func isNotFound(err error) bool {
	var nsk *s3types.NoSuchKey
	if errors.As(err, &nsk) {
		return true
	}

	type apiErr interface{ ErrorCode() string }
	var ae apiErr
	if errors.As(err, &ae) {
		code := ae.ErrorCode()
		if code == "NoSuchKey" || code == "NotFound" {
			return true
		}
	}
	return false
}
