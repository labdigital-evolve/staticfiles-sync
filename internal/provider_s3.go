package internal

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client struct
type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(ctx context.Context, bucket string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &S3Client{client: s3.NewFromConfig(cfg), bucket: bucket}, nil
}

// UploadFile to S3
func (c *S3Client) UploadFile(ctx context.Context, file io.Reader, remotePath string, syncContext FileSyncContext) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(c.bucket),
		Key:          aws.String(remotePath),
		Body:         file,
		CacheControl: aws.String(syncContext.CacheControl),
		ContentType:  aws.String(syncContext.ContentType),
	})
	return err
}

// FileExists in S3
func (c *S3Client) FileExists(ctx context.Context, remotePath string) (bool, error) {
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(remotePath),
	})
	if err != nil {
		var notFoundErr *types.NotFound
		if errors.As(err, &notFoundErr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
