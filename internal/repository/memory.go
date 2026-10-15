package repository

import (
	"context"
	"sync"

	"github.com/JustTony97/url-shortener.git/internal/model"
)

type MemoryRepository struct {
	cache map[string]model.URL
	mu    sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		cache: make(map[string]model.URL),
	}
}

func (r *MemoryRepository) SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[shortenedURL] = model.URL{
		ShortURL:    shortenedURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	return nil
}

func (r *MemoryRepository) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	urlData, ok := r.cache[shortenedURL]
	if !ok {
		return "", false, nil
	}
	return urlData.OriginalURL, true, nil
}

func (r *MemoryRepository) SetShortenedURLs(ctx context.Context, items []model.URL, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range items {
		item.UserID = userID
		r.cache[item.ShortURL] = item
	}
	return nil
}

func (r *MemoryRepository) GetUserURLs(ctx context.Context, userID string) ([]model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userURLs []model.URL

	for _, item := range r.cache {
		if item.UserID == userID {
			userURLs = append(userURLs, item)
		}
	}

	return userURLs, nil
}

func (r *MemoryRepository) DeleteUserURLs(ctx context.Context, userIDs []string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < len(shortURLs); i++ {
		sURL := shortURLs[i]
		uID := userIDs[i]

		if urlData, exists := r.cache[sURL]; exists {
			if urlData.UserID == uID {
				urlData.IsDeleted = true
				r.cache[sURL] = urlData
			}
		}
	}

	return nil
}
