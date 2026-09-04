package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	args := m.Called(ctx, shortenedURL)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockURLService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	args := m.Called(ctx, originalURL)
	return args.String(0), args.Error(1)
}
