package mocks

import "github.com/stretchr/testify/mock"

type MockUrlRepository struct {
	mock.Mock
}

func (m *MockUrlRepository) GetOriginalUrl(shortenedUrl string) (string, bool) {
	args := m.Called(shortenedUrl)
	return args.String(0), args.Bool(1)
}

func (m *MockUrlRepository) SetShortenedUrl(shortenedUrl string, originalUrl string) {
	m.Called(shortenedUrl, originalUrl)
}
