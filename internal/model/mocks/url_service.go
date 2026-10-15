package mocks

import (
	"context"

	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	args := m.Called(ctx, shortenedURL)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockURLService) CreateShortURL(ctx context.Context, originalURL string, userID string) (string, error) {
	args := m.Called(ctx, originalURL, userID)
	return args.String(0), args.Error(1)
}

func (m *MockURLService) CreateShortURLs(ctx context.Context, reqItems []model.BatchItemRequest, userID string) ([]model.BatchResponseItem, error) {
	args := m.Called(ctx, reqItems, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.BatchResponseItem), args.Error(1)
}

func (m *MockURLService) GetUserURLs(ctx context.Context, userID string) ([]model.URL, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.URL), args.Error(1)
}

func (m *MockURLService) DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) {
	m.Called(ctx, userID, shortURLs)
}
