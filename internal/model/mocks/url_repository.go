package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUrlRepository struct {
	mock.Mock
}

func (m *MockUrlRepository) GetOriginalUrl(ctx context.Context, shortenedUrl string) (string, bool, error) {
	args := m.Called(ctx, shortenedUrl)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockUrlRepository) SetShortenedUrl(ctx context.Context, shortenedUrl string, originalUrl string) error {
	args := m.Called(ctx, shortenedUrl, originalUrl)
	return args.Error(0)
}
