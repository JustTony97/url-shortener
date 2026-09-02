package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
)

type UrlRepository interface {
	GetOriginalUrl(ctx context.Context, shortenedUrl string) (string, bool, error)
	SetShortenedUrl(ctx context.Context, shortenedUrl string, originalUrl string) error
}

type UrlService struct {
	repo UrlRepository
}

func NewUrlService(r UrlRepository) *UrlService {
	return &UrlService{repo: r}
}

func (s *UrlService) GetOriginalUrl(ctx context.Context, shortenedUrl string) (string, bool, error) {
	return s.repo.GetOriginalUrl(ctx, shortenedUrl)
}

func (s *UrlService) CreateShortUrl(ctx context.Context, originalUrl string) (string, error) {
	var shortenedUrl string
	var hashSalt = ""

	for i := 0; i < 5; i++ {
		hash := fnv.New32a()
		_, err := hash.Write([]byte(originalUrl + hashSalt))
		if err != nil {
			return "", err
		}
		shortenedUrl = fmt.Sprintf("%x", hash.Sum32())

		url, exists, err := s.repo.GetOriginalUrl(ctx, shortenedUrl)
		if err != nil {
			return "", err
		}

		if !exists {
			err := s.repo.SetShortenedUrl(ctx, shortenedUrl, originalUrl)
			if err != nil {
				return "", err
			}
			return shortenedUrl, nil
		}

		if url == originalUrl {
			return shortenedUrl, nil
		}
		hashSalt = strconv.Itoa(rand.Int())
	}

	return "", errors.New("failed to generate unique short url")
}
