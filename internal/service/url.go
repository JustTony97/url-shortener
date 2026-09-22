package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/model"
)

type URLRepository interface {
	GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error)
	SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string) error
	SetShortenedURLs(ctx context.Context, items []model.URL) error
}

type URLService struct {
	repo URLRepository
	cfg  config.Config
}

func NewURLService(r URLRepository, c config.Config) *URLService {
	return &URLService{repo: r, cfg: c}
}

func (s *URLService) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	return s.repo.GetOriginalURL(ctx, shortenedURL)
}

func (s *URLService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	shortenedURL, err := s.generateShortURL(ctx, originalURL)
	if err != nil {
		return "", err
	}

	err = s.repo.SetShortenedURL(ctx, shortenedURL, originalURL)
	if err != nil {
		return "", err
	}

	return shortenedURL, nil
}

func (s *URLService) CreateShortURLs(ctx context.Context, reqItems []model.BatchItemRequest) ([]model.BatchResponseItem, error) {
	if len(reqItems) == 0 {
		return nil, nil
	}

	responseItems := make([]model.BatchResponseItem, len(reqItems))

	uniqueURLs := make(map[string]model.URL)

	for i, reqItem := range reqItems {
		shortHash, err := s.generateShortURL(ctx, reqItem.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to process URL at index %d: %w", i, err)
		}

		responseItems[i] = model.BatchResponseItem{
			OriginalURL: reqItem.CorrelationID,
			ShortURL:    fmt.Sprintf("%s/%s", s.cfg.BaseURL, shortHash),
		}

		uniqueURLs[shortHash] = model.URL{
			ShortURL:    shortHash,
			OriginalURL: reqItem.OriginalURL,
		}
	}

	dbItems := make([]model.URL, 0, len(uniqueURLs))
	for _, u := range uniqueURLs {
		dbItems = append(dbItems, u)
	}

	err := s.repo.SetShortenedURLs(ctx, dbItems)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	return responseItems, nil
}

func (s *URLService) generateShortURL(ctx context.Context, originalURL string) (string, error) {
	var shortenedURL string
	var hashSalt = ""

	for i := 0; i < 5; i++ {
		hash := fnv.New32a()
		_, err := hash.Write([]byte(originalURL + hashSalt))
		if err != nil {
			return "", err
		}
		shortenedURL = fmt.Sprintf("%x", hash.Sum32())

		url, exists, err := s.repo.GetOriginalURL(ctx, shortenedURL)
		if err != nil {
			return "", err
		}

		if !exists || url == originalURL {
			return shortenedURL, nil
		}

		hashSalt = strconv.Itoa(rand.Int())
	}

	return "", errors.New("failed to generate unique short url due to heavy collisions")
}
