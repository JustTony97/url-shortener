package repository

import (
	"context"
	"sync"

	"github.com/JustTony97/url-shortener.git/internal/model"
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

func (r *MemoryRepository) SetShortenedURLs(ctx context.Context, items []model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range items {
		r.cache[item.ShortURL] = item.OriginalURL
	}

	return nil
}
