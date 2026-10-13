package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/JustTony97/url-shortener.git/internal/config"
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

func (r *FileRepository) SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("file repo: open for reading: %w", err)
	}

	var urls []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&urls); err != nil && err != io.EOF {
		file.Close()
		return fmt.Errorf("file repo: decode json: %w", err)
	}
	file.Close()

	for _, u := range urls {
		if u.OriginalURL == originalURL {
			return model.ErrHashCollision
		}
		if u.ShortURL == shortenedURL {
			return model.ErrConflictURL
		}
	}

	urls = append(urls, model.URL{
		UUID:        "",
		ShortURL:    shortenedURL,
		OriginalURL: originalURL,
		UserID:      userID,
	})

	data, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		return fmt.Errorf("file repo: marshal json: %w", err)
	}

	err = os.WriteFile(r.cfg.FileStoragePath, data, 0644)
	if err != nil {
		return fmt.Errorf("file repo: write file: %w", err)
	}

	return nil
}

func (r *FileRepository) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", false, fmt.Errorf("file repo: open for read: %w", err)
	}
	defer file.Close()

	var urls []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&urls); err != nil {
		if err == io.EOF {
			return "", false, nil
		}
		return "", false, fmt.Errorf("file repo: decode json: %w", err)
	}

	for _, u := range urls {
		if u.ShortURL == shortenedURL {
			return u.OriginalURL, true, nil
		}
	}

	return "", false, nil
}

func (r *FileRepository) SetShortenedURLs(ctx context.Context, items []model.URL, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("file repo: open for batch reading: %w", err)
	}

	var currentURLs []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&currentURLs); err != nil && err != io.EOF {
		file.Close()
		return fmt.Errorf("file repo: decode batch json: %w", err)
	}
	file.Close()

	for _, item := range items {
		item.UserID = userID

		for _, u := range currentURLs {
			if u.OriginalURL == item.OriginalURL {
				return model.ErrHashCollision
			}
			if u.ShortURL == item.ShortURL {
				return model.ErrConflictURL
			}
		}
		currentURLs = append(currentURLs, item)
	}

	data, err := json.MarshalIndent(currentURLs, "", "  ")
	if err != nil {
		return fmt.Errorf("file repo: marshal batch json: %w", err)
	}

	err = os.WriteFile(r.cfg.FileStoragePath, data, 0644)
	if err != nil {
		return fmt.Errorf("file repo: write batch file: %w", err)
	}

	return nil
}

func (r *FileRepository) GetUserURLs(ctx context.Context, userID string) ([]model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, err := os.OpenFile(r.cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("file repo: open for user selection: %w", err)
	}
	defer file.Close()

	var allURLs []model.URL
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&allURLs); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("file repo: decode user urls: %w", err)
	}

	var userURLs []model.URL
	for _, u := range allURLs {
		if u.UserID == userID {
			userURLs = append(userURLs, u)
		}
	}

	return userURLs, nil
}
