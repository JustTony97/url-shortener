package repository

import (
	"context"

	"github.com/JustTony97/url-shortener.git/internal/model"
)

type Repository interface {
	SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string) error
	SetShortenedURLs(ctx context.Context, items []model.URL) error
	GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error)
}
