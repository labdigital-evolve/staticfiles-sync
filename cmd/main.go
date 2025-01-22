package cmd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/labd/cloudstaticfiles/internal"
)

func NewStorageClient(ctx context.Context, provider, bucket string) (internal.StorageClient, error) {
	switch provider {
	case "s3":
		return internal.NewS3Client(ctx, bucket)
	case "gcp":
		return internal.NewGCSClient(ctx, bucket)
	case "azblob":
		return internal.NewAzureBlobClient(ctx, bucket)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

var RootCmd = &cobra.Command{
	Use:   "multicloud-cli",
	Short: "A CLI tool for multi-cloud file operations",
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync a local directory with a specified cloud provider's remote directory",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		localDir, _ := cmd.Flags().GetString("source")
		remoteURL, _ := cmd.Flags().GetString("target")
		lockFile, _ := cmd.Flags().GetString("lockfile")
		cacheControl, _ := cmd.Flags().GetString("cache-control")

		if localDir == "" || remoteURL == "" || lockFile == "" {
			log.Fatalf("source, target and lockfile arguments are required")
		}

		// Parse the remote URL
		parsedURL, err := url.Parse(remoteURL)
		if err != nil {
			log.Fatalf("failed to parse remote URL: %v", err)
		}

		provider := parsedURL.Scheme
		bucket := parsedURL.Host
		remoteDir := strings.TrimPrefix(parsedURL.Path, "/")

		client, err := NewStorageClient(ctx, provider, bucket)
		if err != nil {
			log.Fatalf("failed to initialize storage client: %v", err)
		}

		// Check if lock file exists
		exists, err := client.FileExists(ctx, filepath.Join(remoteDir, lockFile))
		if err != nil {
			log.Fatalf("failed to check lock file existence: %v", err)
		}
		if exists {
			fmt.Println("lock file", lockFile, "exists, skipping sync")
			return
		}

		syncContext := internal.SyncContext{
			CacheControl: cacheControl,
		}

		// Perform the sync
		err = internal.SyncDirectory(ctx, client, localDir, remoteDir, syncContext, 20)
		if err != nil {
			log.Fatalf("directory sync failed: %v", err)
		}
		fmt.Println("Directory synchronized successfully; creating lock file")

		// Create the lockfile (empty file)
		err = client.UploadFile(ctx, strings.NewReader(""), filepath.Join(remoteDir, lockFile), internal.FileSyncContext{})
		if err != nil {
			log.Fatalf("failed to create lock file: %v", err)
		}

		fmt.Println("Lock file", lockFile, "created. Sync process finished.")
	},
}

func init() {
	syncCmd.Flags().StringP("source", "s", "", "Path to the local directory")
	syncCmd.Flags().StringP("target", "t", "", "Remote URL in the format scheme://bucket/path")
	syncCmd.Flags().StringP("lockfile", "l", "", "Remote file path that must not exist before sync")
	syncCmd.Flags().String("cache-control", "", "Cache Control header value for uploaded files")

	syncCmd.MarkFlagRequired("lockfile")
	syncCmd.MarkFlagRequired("source")
	syncCmd.MarkFlagRequired("target")

	RootCmd.AddCommand(syncCmd)
}
