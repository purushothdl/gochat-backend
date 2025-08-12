package s3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types" 
	"github.com/purushothdl/gochat-backend/internal/config"
)

// Client implements the FileStorage interface for AWS S3.
type Client struct {
	s3Client *s3.Client
	bucket   string
	region   string
}

func NewClient(cfg *config.AWSConfig) (*Client, error) {
	sdkConfig, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		s3Client: s3.NewFromConfig(sdkConfig),
		bucket:   cfg.S3Bucket,
		region:   cfg.Region,
	}, nil
}

func (c *Client) Upload(ctx context.Context, key string, contentType string, body io.Reader, isPublic bool) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}
    
	_, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload file to s3: %w", err)
	}
	return nil
}

func (c *Client) Download(ctx context.Context, key string) ([]byte, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from s3: %w", err)
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from s3: %w", err)
	}
	return nil
}

func (c *Client) GetPublicURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucket, c.region, key)
}

func (c *Client) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFoundError *types.NotFound 
		if errors.As(err, &notFoundError) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}
	return true, nil
}
