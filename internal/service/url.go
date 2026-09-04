package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
)

type URLRepository interface {
	GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error)
	SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string) error
}

type URLService struct {
	repo URLRepository
}

func NewURLService(r URLRepository) *URLService {
	return &URLService{repo: r}
}

func (s *URLService) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	return s.repo.GetOriginalURL(ctx, shortenedURL)
}

func (s *URLService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
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

		if !exists {
			err := s.repo.SetShortenedURL(ctx, shortenedURL, originalURL)
			if err != nil {
				return "", err
			}
			return shortenedURL, nil
		}

		if url == originalURL {
			return shortenedURL, nil
		}
		hashSalt = strconv.Itoa(rand.Int())
	}

	return "", errors.New("failed to generate unique short url")
}
