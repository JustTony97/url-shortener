package repository

import "sync"

type UrlRepository struct {
	cache map[string]string
	mu    sync.RWMutex
}

func NewUrlRepository() *UrlRepository {
	return &UrlRepository{
		cache: make(map[string]string),
	}
}

func (r *UrlRepository) SetShortenedUrl(shortenedUrl string, originalUrl string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[shortenedUrl] = originalUrl
}

func (r *UrlRepository) GetOriginalUrl(shortenedUrl string) (string, bool) {
	originalUrl, ok := r.cache[shortenedUrl]
	return originalUrl, ok
}
