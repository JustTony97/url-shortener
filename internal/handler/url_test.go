package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/model/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCfg = config.Config{
	ServerAddr: ":8080",
	BaseUrl:    "http://localhost:8080",
}

func TestUrlHandler_RedirectUrl(t *testing.T) {
	type want struct {
		code           int
		response       string
		contentType    string
		LocationHeader string
	}
	tests := []struct {
		name         string
		want         want
		shortenedUrl string
		setupMock    func(m *mocks.MockUrlService)
	}{
		{
			name:         "positive test",
			shortenedUrl: "4hvjC1",
			want: want{
				code:           http.StatusTemporaryRedirect,
				contentType:    "text/plain",
				response:       "",
				LocationHeader: "https://yandex.ru",
			},
			setupMock: func(m *mocks.MockUrlService) {
				m.On("GetOriginalUrl", "4hvjC1").Return("https://yandex.ru", true)
			},
		},
		{
			name:         "negative test",
			shortenedUrl: "unknown_id",
			want: want{
				code:           http.StatusBadRequest,
				contentType:    "text/plain",
				response:       "Missing original URL",
				LocationHeader: "",
			},
			setupMock: func(m *mocks.MockUrlService) {
				m.On("GetOriginalUrl", "unknown_id").Return("", false)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUrlService)
			tt.setupMock(mockService)

			h := NewUrlHandler(mockService, testCfg)

			r := chi.NewRouter()
			r.Get("/{id}", h.RedirectUrl)

			request := httptest.NewRequest(http.MethodGet, "/"+tt.shortenedUrl, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.name == "positive test" {
				assert.Equal(t, tt.want.LocationHeader, res.Header.Get("Location"))
			}

			bodyBytes, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.want.response != "" {
				assert.Contains(t, string(bodyBytes), tt.want.response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUrlHandler_ShortUrl(t *testing.T) {
	type want struct {
		code           int
		contentType    string
		expectedPrefix string
	}

	tests := []struct {
		name        string
		body        string
		contentType string
		want        want
		setupMock   func(m *mocks.MockUrlService)
	}{
		{
			name:        "positive test",
			body:        "https://yandex.ru",
			contentType: "text/plain",
			want: want{
				code:           http.StatusCreated,
				contentType:    "text/plain",
				expectedPrefix: testCfg.BaseUrl + "/4hvjC1",
			},
			setupMock: func(m *mocks.MockUrlService) {
				m.On("CreateShortUrl", "https://yandex.ru").Return("4hvjC1", nil)
			},
		},
		{
			name:        "negative test - empty body",
			body:        "",
			contentType: "text/plain",
			want: want{
				code:           http.StatusBadRequest,
				contentType:    "text/plain; charset=utf-8",
				expectedPrefix: "Body is empty",
			},
			setupMock: func(m *mocks.MockUrlService) {},
		},
		{
			name:        "negative test - unsupported content-type",
			body:        "https://yandex.ru",
			contentType: "application/json",
			want: want{
				code:           http.StatusBadRequest,
				contentType:    "text/plain; charset=utf-8",
				expectedPrefix: "Unsupported content-type",
			},
			setupMock: func(m *mocks.MockUrlService) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			mockService := new(mocks.MockUrlService)
			tt.setupMock(mockService)

			h := NewUrlHandler(mockService, testCfg)

			h.ShortUrl(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)

			bodyBytes, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			responseString := string(bodyBytes)

			assert.Contains(t, responseString, tt.want.expectedPrefix)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUrlHandler_ShortenUrl(t *testing.T) {
	type want struct {
		code           int
		contentType    string
		expectedPrefix string
	}

	tests := []struct {
		name        string
		body        string
		contentType string
		want        want
		setupMock   func(m *mocks.MockUrlService)
	}{
		{
			name:        "positive test",
			body:        "{\"url\": \"https://yandex.ru\"}",
			contentType: "application/json",
			want: want{
				code:           http.StatusCreated,
				contentType:    "application/json",
				expectedPrefix: testCfg.BaseUrl + "/4hvjC1",
			},
			setupMock: func(m *mocks.MockUrlService) {
				m.On("CreateShortUrl", "https://yandex.ru").Return("4hvjC1", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))

			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			mockService := new(mocks.MockUrlService)
			tt.setupMock(mockService)

			h := NewUrlHandler(mockService, testCfg)

			h.ShortenUrl(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)

			bodyBytes, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			responseString := string(bodyBytes)

			assert.Contains(t, responseString, tt.want.expectedPrefix)

			mockService.AssertExpectations(t)
		})
	}
}
