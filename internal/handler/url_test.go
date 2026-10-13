package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/middlewares"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/JustTony97/url-shortener.git/internal/model/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testCfg = config.Config{
	ServerAddr: ":8080",
	BaseURL:    "http://localhost:8080",
}

func TestURLHandler_RedirectURL(t *testing.T) {
	type want struct {
		code           int
		response       string
		contentType    string
		LocationHeader string
	}
	tests := []struct {
		name         string
		want         want
		shortenedURL string
		setupMock    func(m *mocks.MockURLService)
	}{
		{
			name:         "positive test",
			shortenedURL: "4hvjC1",
			want: want{
				code:           http.StatusTemporaryRedirect,
				contentType:    "text/plain",
				response:       "",
				LocationHeader: "https://yandex.ru",
			},
			setupMock: func(m *mocks.MockURLService) {
				m.On("GetOriginalURL", mock.Anything, "4hvjC1").Return("https://yandex.ru", true, nil)
			},
		},
		{
			name:         "negative test",
			shortenedURL: "unknown_id",
			want: want{
				code:           http.StatusBadRequest,
				contentType:    "text/plain",
				response:       "Missing original URL",
				LocationHeader: "",
			},
			setupMock: func(m *mocks.MockURLService) {
				m.On("GetOriginalURL", mock.Anything, "unknown_id").Return("", false, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockURLService)
			tt.setupMock(mockService)

			h := NewURLHandler(mockService, testCfg)

			r := chi.NewRouter()
			r.Get("/{id}", h.RedirectURL)

			request := httptest.NewRequest(http.MethodGet, "/"+tt.shortenedURL, nil)
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

func TestURLHandler_ShortURL(t *testing.T) {
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
		setupMock   func(m *mocks.MockURLService)
	}{
		{
			name:        "positive test",
			body:        "https://yandex.ru",
			contentType: "text/plain",
			want: want{
				code:           http.StatusCreated,
				contentType:    "text/plain",
				expectedPrefix: testCfg.BaseURL + "/4hvjC1",
			},
			setupMock: func(m *mocks.MockURLService) {
				m.On("CreateShortURL", mock.Anything, "https://yandex.ru", mock.Anything).Return("4hvjC1", nil)
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
			setupMock: func(m *mocks.MockURLService) {},
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
			setupMock: func(m *mocks.MockURLService) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			ctx := middlewares.ContextWithUserID(request.Context(), "test_user")
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()

			mockService := new(mocks.MockURLService)
			tt.setupMock(mockService)

			h := NewURLHandler(mockService, testCfg)
			h.ShortURL(w, request)

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

func TestURLHandler_ShortenURL(t *testing.T) {
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
		setupMock   func(m *mocks.MockURLService)
	}{
		{
			name:        "positive test",
			body:        "{\"url\": \"https://yandex.ru\"}",
			contentType: "application/json",
			want: want{
				code:           http.StatusCreated,
				contentType:    "application/json",
				expectedPrefix: testCfg.BaseURL + "/4hvjC1",
			},
			setupMock: func(m *mocks.MockURLService) {
				m.On("CreateShortURL", mock.Anything, "https://yandex.ru", mock.Anything).Return("4hvjC1", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))

			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			ctx := middlewares.ContextWithUserID(request.Context(), "test_user")
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()

			mockService := new(mocks.MockURLService)
			tt.setupMock(mockService)

			h := NewURLHandler(mockService, testCfg)
			h.ShortenURL(w, request)

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

func TestURLHandler_GetUserURLs(t *testing.T) {
	tests := []struct {
		name         string
		setAuth      bool
		userID       string
		mockData     []model.URL
		expectedCode int
		validateJSON bool
	}{
		{
			name:    "success case with records",
			setAuth: true,
			userID:  "user_123",
			mockData: []model.URL{
				{OriginalURL: "https://yandex.ru", ShortURL: "http://localhost:8080/4hvjC1"},
			},
			expectedCode: http.StatusOK,
			validateJSON: true,
		},
		{
			name:         "success case empty list",
			setAuth:      true,
			userID:       "user_empty",
			mockData:     []model.URL{},
			expectedCode: http.StatusNoContent,
			validateJSON: false,
		},
		{
			name:         "unauthorized case",
			setAuth:      false,
			expectedCode: http.StatusUnauthorized,
			validateJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockURLService)
			if tt.setAuth {
				mockService.On("GetUserURLs", mock.Anything, tt.userID).Return(tt.mockData, nil)
			}

			h := NewURLHandler(mockService, testCfg)

			request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			if tt.setAuth {
				ctx := middlewares.ContextWithUserID(request.Context(), tt.userID)
				request = request.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			h.GetUserURLs(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedCode, res.StatusCode)

			if tt.validateJSON {
				assert.Contains(t, res.Header.Get("Content-Type"), "application/json")
				var resp model.APIUserShortenBatchResponse
				err := json.NewDecoder(res.Body).Decode(&resp)
				require.NoError(t, err)
				require.Len(t, resp, len(tt.mockData))

				assert.Equal(t, tt.mockData[0].OriginalURL, resp[0].OriginalURL)
				assert.Equal(t, tt.mockData[0].ShortURL, resp[0].ShortURL)
			}
		})
	}
}
