package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type SpacesUploader struct {
	s3            *s3.Client
	bucket        string
	publicBaseURL string
}

type SpacesConfig struct {
	Endpoint      string
	Region        string
	Bucket        string
	AccessKey     string
	SecretKey     string
	PublicBaseURL string
}

func NewSpacesUploader(ctx context.Context, cfg SpacesConfig) (*SpacesUploader, error) {
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint) 
	})

	return &SpacesUploader{
		s3:            client,
		bucket:        cfg.Bucket,
		publicBaseURL: cfg.PublicBaseURL,
	}, nil
}

func (u *SpacesUploader) PutJPEG(ctx context.Context, key string, body []byte) (string, error) {
	_, err := u.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(u.bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(body),
		ContentType:  aws.String("image/jpeg"),
		CacheControl: aws.String("public, max-age=31536000, immutable"), // one year caching 
		ACL:          types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", err
	}
	return u.publicBaseURL + "/" + key, nil
}