package repository

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	cache map[string]string
	mu    sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		cache: make(map[string]string),
	}
}

func (r *MemoryRepository) SetShortenedUrl(ctx context.Context, shortenedUrl string, originalUrl string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[shortenedUrl] = originalUrl
	return nil
}

func (r *MemoryRepository) GetOriginalUrl(ctx context.Context, shortenedUrl string) (string, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	originalUrl, ok := r.cache[shortenedUrl]
	return originalUrl, ok, nil
}
