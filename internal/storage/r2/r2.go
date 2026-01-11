// internal/storage/r2/r2.go
package r2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/zetaoss/zavatar/internal/storage"
)

type Storage struct {
	client     *s3.Client
	bucket     string
	publicBase string
}

func New(accountID, bucket, accessKey, secretKey, publicBase string) (*Storage, error) {
	if publicBase == "" {
		return nil, errors.New("r2: missing public base")
	}

	ep := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(ep)
	})

	publicBase = strings.TrimRight(publicBase, "/")
	if !strings.Contains(publicBase, "://") {
		publicBase = "https://" + publicBase
	}

	return &Storage{
		client:     client,
		bucket:     bucket,
		publicBase: publicBase,
	}, nil
}

func (s *Storage) Put(ctx context.Context, key string, contentType string, body []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}

	var nsk *types.NotFound
	if errors.As(err, &nsk) {
		return false, nil
	}

	var sk *types.NoSuchKey
	if errors.As(err, &sk) {
		return false, nil
	}

	return false, err
}

func (s *Storage) PublicURL(key string) (string, error) {
	return s.publicBase + "/" + key, nil
}

var _ storage.Storage = (*Storage)(nil)
