package service

import (
	"fmt"
	"hash/fnv"

	"github.com/JustTony97/url-shortener.git/internal/repository"
)

type UrlService struct {
	repo *repository.UrlRepository
}

func NewUrlService(r *repository.UrlRepository) *UrlService {
	return &UrlService{repo: r}
}

func (s *UrlService) GetOriginalUrl(shortedUrl string) (string, bool) {
	return s.repo.GetOriginalUrl(shortedUrl)
}

func (s *UrlService) CreateShortUrl(originalUrl string) string {
	hash := fnv.New32a()
	hash.Write([]byte(originalUrl))
	shortedUrl := fmt.Sprintf("%x", hash.Sum32())

	s.repo.SetShortedUrl(shortedUrl, originalUrl)
	return shortedUrl
}
