package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUrlService struct {
	mock.Mock
}

func (m *MockUrlService) GetOriginalUrl(ctx context.Context, shortenedUrl string) (string, bool, error) {
	args := m.Called(ctx, shortenedUrl)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockUrlService) CreateShortUrl(ctx context.Context, originalUrl string) (string, error) {
	args := m.Called(ctx, originalUrl)
	return args.String(0), args.Error(1)
}
