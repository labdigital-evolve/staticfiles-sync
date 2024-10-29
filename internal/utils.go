package internal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type StorageClient interface {
	UploadFile(ctx context.Context, file io.Reader, remotePath string) error
	FileExists(ctx context.Context, remotePath string) (bool, error)
}

func SyncDirectory(ctx context.Context, client StorageClient, localDir, remoteDir string, concurrency int) error {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)

	_ = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, _ := filepath.Rel(localDir, path)
		remotePath := filepath.Join(remoteDir, relPath)

		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			exists, err := client.FileExists(ctx, remotePath)
			if err != nil {
				errCh <- err
				return
			}
			if !exists {
				file, err := os.Open(path)
				if err != nil {
					errCh <- err
					return
				}
				defer file.Close()
				if err := client.UploadFile(ctx, file, remotePath); err != nil {
					errCh <- err
				} else {
					fmt.Printf("Uploaded %s to %s\n", path, remotePath)
				}
			} else {
				fmt.Printf("Skipping existing file: %s\n", remotePath)
			}
		}()
		return nil
	})

	// Wait for all goroutines and close the error channel
	wg.Wait()
	close(errCh)

	// Check errors from channel
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
