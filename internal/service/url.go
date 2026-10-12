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
	var hashSalt string
	const attempts = 5

	for a := 0; a < attempts; a++ {
		shortenedURL, err := s.generateShortURL(originalURL, hashSalt)
		if err != nil {
			return "", err
		}

		err = s.repo.SetShortenedURL(ctx, shortenedURL, originalURL)
		if err != nil {
			if errors.Is(err, model.ErrHashCollision) {
				hashSalt = strconv.Itoa(rand.Int())
				continue
			}
			if errors.Is(err, model.ErrConflictURL) {
				return shortenedURL, &model.ErrConflictWithExistingURL{ShortURL: shortenedURL}
			}
			return "", err
		}

		return shortenedURL, nil
	}

	return "", errors.New("failed to generate unique short url due to heavy collisions")
}

func (s *URLService) CreateShortURLs(ctx context.Context, reqItems []model.BatchItemRequest) ([]model.BatchResponseItem, error) {
	if len(reqItems) == 0 {
		return nil, nil
	}

	responseItems := make([]model.BatchResponseItem, len(reqItems))
	uniqueURLs := make(map[string]model.URL)
	origToHash := make(map[string]string)
	hashToOrig := make(map[string]string)

	for i, reqItem := range reqItems {
		shortHash, err := s.resolveShortURL(reqItem.OriginalURL, origToHash, hashToOrig)
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

	if err := s.repo.SetShortenedURLs(ctx, dbItems); err != nil {
		return nil, err
	}

	return responseItems, nil
}

func (s *URLService) resolveShortURL(originalURL string, origToHash, hashToOrig map[string]string) (string, error) {
	if existingHash, exists := origToHash[originalURL]; exists {
		return existingHash, nil
	}

	var hashSalt string
	const attempts = 5

	for a := 0; a < attempts; a++ {
		shortHash, err := s.generateShortURL(originalURL, hashSalt)
		if err != nil {
			return "", err
		}

		if _, existsHash := hashToOrig[shortHash]; existsHash {
			hashSalt = strconv.Itoa(rand.Int())
			continue
		}

		origToHash[originalURL] = shortHash
		hashToOrig[shortHash] = originalURL
		return shortHash, nil
	}

	return "", fmt.Errorf("failed to generate unique URL after %d attempts", attempts)
}

func (s *URLService) generateShortURL(originalURL string, hashSalt string) (string, error) {
	hash := fnv.New32a()
	_, err := hash.Write([]byte(originalURL + hashSalt))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum32()), nil
}
