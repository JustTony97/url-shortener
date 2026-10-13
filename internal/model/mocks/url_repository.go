package mocks

import (
	"context"

	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	args := m.Called(ctx, shortenedURL)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockURLRepository) SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string, userID string) error {
	args := m.Called(ctx, shortenedURL, originalURL, userID)
	return args.Error(0)
}

func (m *MockURLRepository) SetShortenedURLs(ctx context.Context, items []model.URL, userID string) error {
	args := m.Called(ctx, items, userID)
	return args.Error(0)
}

func (m *MockURLRepository) GetUserURLs(ctx context.Context, userID string) ([]model.URL, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.URL), args.Error(1)
}
