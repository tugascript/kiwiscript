package services

import (
	"context"
	"github.com/google/uuid"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	objStg "github.com/kiwiscript/kiwiscript_go/providers/object_storage"
	"sync"
)

type FindFileURLOptions struct {
	UserID  int32
	FileID  uuid.UUID
	FileExt string
}

func (s *Services) FindFileURL(ctx context.Context, opts FindFileURLOptions) (string, *ServiceError) {
	log := s.log.WithGroup("services.files.FindFileURL").With(
		"userId", opts.UserID,
		"fileId", opts.FileID,
		"fileExt", opts.FileExt,
	)
	log.Info("Finding file URL...")

	cacheOpts := cc.GetFileURLOptions{
		UserID: opts.UserID,
		FileID: opts.FileID,
	}
	if url, err := s.cache.GetFileURL(cacheOpts); err == nil {
		return url, nil
	}

	url, err := s.objStg.GetFileUrl(ctx, objStg.GetFileURLOptions{
		UserID:  opts.UserID,
		FileID:  opts.FileID,
		FileExt: opts.FileExt,
	})
	if err != nil {
		log.Error("Error getting file URL", "error", err)
		return "", NewServerError("Error getting file URL")
	}

	return url, nil
}

type FileURLsContainer struct {
	mu   sync.Mutex
	urls map[uuid.UUID]string
}

func (c *FileURLsContainer) Set(fileID uuid.UUID, url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.urls[fileID] = url
}

func (c *FileURLsContainer) Get(fileID uuid.UUID) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	url, ok := c.urls[fileID]
	return url, ok
}

func (s *Services) FindFileURLs(ctx context.Context, opts []FindFileURLOptions) (*FileURLsContainer, *ServiceError) {
	var wg sync.WaitGroup
	container := FileURLsContainer{
		urls: make(map[uuid.UUID]string),
	}
	errChan := make(chan *ServiceError)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, opt := range opts {
		wg.Add(1)
		go func(opt FindFileURLOptions) {
			defer wg.Done()
			url, err := s.FindFileURL(ctx, opt)
			if err != nil {
				errChan <- err
				cancel() // Cancel remaining goroutines
				return
			}
			container.Set(opt.FileID, url)
		}(opt)
	}
	wg.Wait()

	select {
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return &container, nil
	}
}
