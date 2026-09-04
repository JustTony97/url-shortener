package repository

import "context"

type Repository interface {
	SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string) error
	GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error)
}
