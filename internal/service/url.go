package service

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
)

type UrlRepository interface {
	GetOriginalUrl(shortenedUrl string) (string, bool)
	SetShortenedUrl(shortenedUrl string, originalUrl string)
}

type UrlService struct {
	repo UrlRepository
}

func NewUrlService(r UrlRepository) *UrlService {
	return &UrlService{repo: r}
}

func (s *UrlService) GetOriginalUrl(shortenedUrl string) (string, bool) {
	return s.repo.GetOriginalUrl(shortenedUrl)
}

func (s *UrlService) CreateShortUrl(originalUrl string) (string, error) {
	var shortenedUrl string
	var hashSalt = ""

	for i := 0; i < 5; i++ {
		hash := fnv.New32a()
		hash.Write([]byte(originalUrl + hashSalt))
		shortenedUrl = fmt.Sprintf("%x", hash.Sum32())

		url, exists := s.repo.GetOriginalUrl(shortenedUrl)

		if !exists {
			s.repo.SetShortenedUrl(shortenedUrl, originalUrl)
			return shortenedUrl, nil
		}

		if url == originalUrl {
			return shortenedUrl, nil
		}
		hashSalt = strconv.Itoa(rand.Int())
	}

	return "", errors.New("Failed to generate unique short url")
}
