package repository

import "context"

type Repository interface {
	SetShortenedUrl(ctx context.Context, shortenedUrl string, originalUrl string) error
	GetOriginalUrl(ctx context.Context, shortenedUrl string) (string, bool, error)
}
