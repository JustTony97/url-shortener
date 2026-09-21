package repository

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/model"
)

type FileRepository struct {
	cfg config.Config
	mu  sync.RWMutex
}

func NewFileRepository(cfg config.Config) *FileRepository {
	return &FileRepository{
		cfg: cfg,
	}
}

func (r *FileRepository) SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Errorf("failed to open file for reading: %v", err)
		return err
	}

	var urls []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&urls); err != nil && err != io.EOF {
		file.Close()
		logger.Log.Errorf("failed to decode urls: %v", err)
		return err
	}
	file.Close()

	found := false
	for i, u := range urls {
		if u.ShortURL == shortenedURL {
			urls[i].OriginalURL = originalURL
			found = true
			break
		}
	}

	if !found {
		urls = append(urls, model.URL{
			UUID:        "",
			ShortURL:    shortenedURL,
			OriginalURL: originalURL,
		})
	}

	data, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		logger.Log.Errorf("failed to marshal urls: %v", err)
		return err
	}

	err = os.WriteFile(r.cfg.FileStoragePath, data, 0644)
	if err != nil {
		logger.Log.Errorf("failed to write urls to file %s: %v", r.cfg.FileStoragePath, err)
		return err
	}

	return nil
}

func (r *FileRepository) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Errorf("failed to open file: %v", err)
		return "", false, err
	}
	defer file.Close()

	var urls []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&urls); err != nil {
		if err == io.EOF {
			return "", false, nil
		}
		logger.Log.Errorf("failed to decode urls: %v", err)
		return "", false, err
	}

	for _, u := range urls {
		if u.ShortURL == shortenedURL {
			return u.OriginalURL, true, nil
		}
	}

	return "", false, nil
}

func (r *FileRepository) SetShortenedURLs(ctx context.Context, items []model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Errorf("failed to open file for reading batch: %v", err)
		return err
	}

	var currentURLs []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&currentURLs); err != nil && err != io.EOF {
		file.Close()
		logger.Log.Errorf("failed to decode urls during batch: %v", err)
		return err
	}
	file.Close()

	batchMap := make(map[string]model.URL, len(items))
	for _, item := range items {
		batchMap[item.ShortURL] = item
	}

	for i, u := range currentURLs {
		if newItem, exists := batchMap[u.ShortURL]; exists {
			currentURLs[i].OriginalURL = newItem.OriginalURL
			if newItem.UUID != "" {
				currentURLs[i].UUID = newItem.UUID
			}
			delete(batchMap, u.ShortURL)
		}
	}

	for _, newItem := range batchMap {
		currentURLs = append(currentURLs, newItem)
	}

	data, err := json.MarshalIndent(currentURLs, "", "  ")
	if err != nil {
		logger.Log.Errorf("failed to marshal urls during batch: %v", err)
		return err
	}

	err = os.WriteFile(r.cfg.FileStoragePath, data, 0644)
	if err != nil {
		logger.Log.Errorf("failed to write urls to file %s during batch: %v", r.cfg.FileStoragePath, err)
		return err
	}

	return nil
}
