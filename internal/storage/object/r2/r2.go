// internal/storage/object/r2/r2.go
package r2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	AccountID       string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string
	PublicBase      string
}

type Store struct {
	bucket     string
	prefix     string
	publicBase string
	s3         *s3.Client
}

func New(ctx context.Context, c Config) (*Store, error) {
	if c.AccountID == "" || c.Bucket == "" || c.AccessKeyID == "" || c.SecretAccessKey == "" {
		return nil, fmt.Errorf("r2: missing required config")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.AccountID)
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("r2: invalid endpoint: %w", err)
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("r2: load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &Store{
		bucket:     c.Bucket,
		prefix:     c.Prefix,
		publicBase: strings.TrimRight(c.PublicBase, "/"),
		s3:         client,
	}, nil
}

func (s *Store) key(k string) string {
	if s.prefix == "" {
		return k
	}
	return s.prefix + k
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(key)),
	})
	if err == nil {
		return true, nil
	}

	var nf *s3types.NotFound
	if errors.As(err, &nf) {
		return false, nil
	}

	return false, err
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	out, err := s.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(key)),
	})
	if err != nil {
		return nil, "", err
	}

	ct := aws.ToString(out.ContentType)
	return out.Body, ct, nil
}

func (s *Store) Put(ctx context.Context, key string, contentType string, body []byte) error {
	_, err := s.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.key(key)),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *Store) PublicURL(key string) string {
	if s.publicBase == "" {
		return ""
	}
	return s.publicBase + "/" + s.key(key)
}
