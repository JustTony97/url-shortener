package handler

import (
	"fmt"
	"io"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/service"
	"github.com/go-chi/chi"

	"net/http"
	"strings"
)

type UrlHandler struct {
	service *service.UrlService
}

func NewUrlHandler(s *service.UrlService) *UrlHandler {
	return &UrlHandler{service: s}
}

func (h *UrlHandler) RedirectUrl(w http.ResponseWriter, r *http.Request) {
	shortedUrl := chi.URLParam(r, "id")
	if shortedUrl == "" {
		http.Error(w, "ID param is missing", http.StatusBadRequest)
		return
	}

	originalUrl, ok := h.service.GetOriginalUrl(shortedUrl)
	if !ok {
		http.Error(w, "Missing original URL", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
}

func (h *UrlHandler) ShortUrl(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
		http.Error(w, "Unsupported content-type", http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalUrl := string(bodyBytes)
	if originalUrl == "" {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	shortedUrl := h.service.CreateShortUrl(originalUrl)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fullShortedURL := fmt.Sprint(config.RedirectBaseUrl + "/" + shortedUrl)
	w.Write([]byte(fullShortedURL))
}
