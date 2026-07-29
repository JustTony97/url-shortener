package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/repository"
	"github.com/JustTony97/url-shortener.git/internal/service"
	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewUrlRepository()
			srv := service.NewUrlService(repo)
			h := NewUrlHandler(srv)

			if tt.name == "positive test" {
				repo.SetShortenedUrl(tt.shortenedUrl, tt.want.LocationHeader)
			}

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
	}{
		{
			name:        "positive test",
			body:        "https://yandex.ru",
			contentType: "text/plain",
			want: want{
				code:           http.StatusCreated,
				contentType:    "text/plain",
				expectedPrefix: config.RedirectBaseUrl,
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			repo := repository.NewUrlRepository()
			srv := service.NewUrlService(repo)
			h := NewUrlHandler(srv)

			h.ShortUrl(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)

			bodyBytes, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			responseString := string(bodyBytes)

			assert.Contains(t, responseString, tt.want.expectedPrefix)

			if tt.want.code == http.StatusCreated {
				shortenedUrl := strings.TrimPrefix(responseString, tt.want.expectedPrefix)

				shortenedUrl = strings.TrimPrefix(shortenedUrl, "/")
				shortenedUrl = strings.TrimSpace(shortenedUrl)

				savedUrl, exists := repo.GetOriginalUrl(shortenedUrl)
				assert.True(t, exists)
				assert.Equal(t, tt.body, savedUrl)
			}
		})
	}
}
