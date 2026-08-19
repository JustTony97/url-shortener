package repository

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/model"
)

type UrlRepository struct {
	cache map[string]string
	cfg   config.Config
	mu    sync.RWMutex
}

func NewUrlRepository(cfg config.Config) *UrlRepository {
	return &UrlRepository{
		cache: make(map[string]string),
		cfg:   cfg,
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

func (r *UrlRepository) Load() error {
	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var urls []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&urls); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	for _, url := range urls {
		r.SetShortenedUrl(url.ShortUrl, url.OriginalUrl)
	}
	logger.Log.Infof("Succesfully loaded %d urls!", len(r.cache))
	return nil
}

func (r *UrlRepository) Flush() error {
	r.mu.RLock()
	urls := make([]model.URL, 0, len(r.cache))
	for short, original := range r.cache {
		urls = append(urls, model.URL{
			UUID:        "",
			ShortUrl:    short,
			OriginalUrl: original,
		})
	}
	r.mu.RUnlock()

	data, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.cfg.FileStoragePath, data, 0644)
}
