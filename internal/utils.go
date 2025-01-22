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
	UploadFile(ctx context.Context, file io.Reader, remotePath string, syncContext FileSyncContext) error
	FileExists(ctx context.Context, remotePath string) (bool, error)
}

type SyncContext struct {
	CacheControl string
}

type FileSyncContext struct {
	CacheControl string
	ContentType  string
}

func SyncDirectory(ctx context.Context, client StorageClient, localDir, remoteDir string, syncContext SyncContext, concurrency int) error {
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

				// Get the content type based on file extension.
				// We can't rely on auto-detection of mimetypes because
				// of security implications. See https://stackoverflow.com/questions/70695214/net-http-does-detectcontenttype-support-javascript
				contentType := GetContentType(path)

				// Use input from syncContext to create the FileSyncContext in combiation with contentType
				fileSyncContext := FileSyncContext{
					CacheControl: syncContext.CacheControl,
					ContentType:  contentType,
				}

				if err := client.UploadFile(ctx, file, remotePath, fileSyncContext); err != nil {
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

func GetContentType(file string) string {
	ext := filepath.Ext(file)
	switch ext {
	case ".htm", ".html":
		return "text/html; charset=UTF-8"
	case ".css":
		return "text/css; charset=UTF-8"
	case ".js":
		return "application/javascript; charset=UTF-8"
	}

	return ""
}
