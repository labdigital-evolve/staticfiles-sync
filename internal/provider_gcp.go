package internal

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

// GCSClient struct
type GCSClient struct {
	client *storage.Client
	bucket string
}

func NewGCSClient(ctx context.Context, bucket string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GCSClient{client: client, bucket: bucket}, nil
}

// UploadFile to GCS
func (c *GCSClient) UploadFile(ctx context.Context, file io.Reader, remotePath string, contentType string) error {
	bucket := c.client.Bucket(c.bucket)
	obj := bucket.Object(remotePath)

	writer := obj.NewWriter(ctx)

	if contentType != "" {
		writer.ContentType = contentType
	}

	defer writer.Close()

	_, err := io.Copy(writer, file)
	return err
}

// FileExists in GCS
func (c *GCSClient) FileExists(ctx context.Context, remotePath string) (bool, error) {
	_, err := c.client.Bucket(c.bucket).Object(remotePath).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	return err == nil, err
}
