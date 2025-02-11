package internal

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// AzureBlobClient struct
type AzureBlobClient struct {
	client    *azblob.Client
	container string
}

// NewAzureBlobClient initializes a new AzureBlobClient
func NewAzureBlobClient(ctx context.Context, bucket string) (*AzureBlobClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	parts := strings.SplitN(bucket, "-", 2)
	accountName := parts[0]
	container := parts[1]

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	client, err := azblob.NewClient(serviceURL, cred, nil)
	if err != nil {
		return nil, err
	}
	return &AzureBlobClient{client: client, container: container}, nil
}

func (c *AzureBlobClient) UploadFile(ctx context.Context, file io.Reader, remotePath string, syncContext FileSyncContext) error {
	var tempFile *os.File
	var err error

	// Check if file is *os.File
	if f, ok := file.(*os.File); ok {
		tempFile = f
	} else {
		// Write to temporary file if not *os.File
		tempFile, err = ioutil.TempFile("", "temp-upload")
		if err != nil {
			return err
		}
		defer os.Remove(tempFile.Name())

		_, err = io.Copy(tempFile, file)
		if err != nil {
			return err
		}

		// Reset the file pointer to the beginning
		if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}
	defer tempFile.Close()

	// Upload directly using the client and specifying the container and blob name
	_, err = c.client.UploadFile(ctx, c.container, remotePath, tempFile, &azblob.UploadFileOptions{
		BlockSize:   4 * 1024 * 1024, // 4MB block size
		Concurrency: 16,              // Adjust concurrency as needed
	})
	return err
}

// FileExists checks if a file exists in Azure Blob storage
func (c *AzureBlobClient) FileExists(ctx context.Context, remotePath string) (bool, error) {
	// Get the container client
	containerClient := c.client.ServiceClient().NewContainerClient(c.container)

	// Get the blob client within the container for the specific file
	blobClient := containerClient.NewBlockBlobClient(remotePath)

	// Attempt to retrieve blob properties to check existence
	_, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		fmt.Println(err)
		return false, nil // Blob does not exist
	}
	return true, nil // Blob exists
}
