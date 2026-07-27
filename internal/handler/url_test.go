package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JustTony97/url-shortener.git/internal/repository"
	"github.com/JustTony97/url-shortener.git/internal/service"
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
		name       string
		body       string
		want       want
		shortedUrl string
	}{
		{
			name:       "positive test",
			shortedUrl: "4hvjC1",

			want: want{
				code:           http.StatusTemporaryRedirect,
				contentType:    "text/plain",
				response:       "",
				LocationHeader: "https://yandex.ru",
			},
		},
		{
			name:       "negative test",
			shortedUrl: "4hvjC1",
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
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("id", tt.shortedUrl)

			w := httptest.NewRecorder()

			repo := repository.NewUrlRepository()
			srv := service.NewUrlService(repo)
			h := NewUrlHandler(srv)

			if tt.name == "positive test" {
				repo.SetShortedUrl(tt.shortedUrl, tt.want.LocationHeader)
			}

			h.RedirectUrl(w, request)
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
