package services

import (
	"context"
	"github.com/google/uuid"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	objStg "github.com/kiwiscript/kiwiscript_go/providers/object_storage"
	"sync"
)

const filesLocation string = "files"

type FindFileURLOptions struct {
	RequestID string
	UserID    int32
	FileID    uuid.UUID
	FileExt   string
}

func (s *Services) FindFileURL(ctx context.Context, opts FindFileURLOptions) (string, *ServiceError) {
	log := s.buildLogger(opts.RequestID, filesLocation, "FindFileURL").With(
		"userId", opts.UserID,
		"fileId", opts.FileID,
		"fileExt", opts.FileExt,
	)
	log.Info("Finding file URL...")

	cacheOpts := cc.GetFileURLOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
		FileID:    opts.FileID,
	}
	if url, err := s.cache.GetFileURL(ctx, cacheOpts); err == nil && url != "" {
		return url, nil
	}

	url, err := s.objStg.GetFileUrl(ctx, objStg.GetFileURLOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
		FileID:    opts.FileID,
		FileExt:   opts.FileExt,
	})
	if err != nil {
		log.Error("Error getting file URL", "error", err)
		return "", NewServerError()
	}

	addFileUrlOpts := cc.AddFileURLOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
		FileID:    opts.FileID,
		URL:       url,
	}
	if err := s.cache.AddFileURL(ctx, addFileUrlOpts); err != nil {
		log.Error("Error caching file URL", "error", err)
		return "", NewServerError()
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, opt := range opts {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if ctx.Err() != nil {
				return
			}

			url, err := s.FindFileURL(ctx, opt)
			if err != nil {
				cancel()
				return
			}

			container.Set(opt.FileID, url)
		}()
	}
	wg.Wait()

	if ctx.Err() != nil {
		return nil, NewServerError()
	}

	return &container, nil
}
