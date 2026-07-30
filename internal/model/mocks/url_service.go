package mocks

import "github.com/stretchr/testify/mock"

type MockUrlService struct {
	mock.Mock
}

func (m *MockUrlService) GetOriginalUrl(shortenedUrl string) (string, bool) {
	args := m.Called(shortenedUrl)
	return args.String(0), args.Bool(1)
}

func (m *MockUrlService) CreateShortUrl(originalUrl string) (string, error) {
	args := m.Called(originalUrl)
	return args.String(0), args.Error(1)
}
