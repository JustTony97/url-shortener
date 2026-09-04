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

func (r *MemoryRepository) SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[shortenedURL] = originalURL
	return nil
}

func (r *MemoryRepository) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	originalURL, ok := r.cache[shortenedURL]
	return originalURL, ok, nil
}
