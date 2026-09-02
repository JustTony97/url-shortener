package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/go-chi/chi/v5"

	"net/http"
	"strings"
)

type UrlHandler struct {
	service UrlService
	cfg     config.Config
}

type UrlService interface {
	GetOriginalUrl(ctx context.Context, shortenedUrl string) (string, bool, error)
	CreateShortUrl(ctx context.Context, originalUrl string) (string, error)
}

func NewUrlHandler(s UrlService, c config.Config) *UrlHandler {
	return &UrlHandler{service: s, cfg: c}
}

func (h *UrlHandler) RedirectUrl(w http.ResponseWriter, r *http.Request) {
	shortenedUrl := chi.URLParam(r, "id")
	if shortenedUrl == "" {
		http.Error(w, "ID param is missing", http.StatusBadRequest)
		return
	}

	originalUrl, ok, err := h.service.GetOriginalUrl(r.Context(), shortenedUrl)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
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

	shortenedUrl, err := h.service.CreateShortUrl(r.Context(), originalUrl)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fullShortenedUrl := fmt.Sprintf("%s/%s", h.cfg.BaseUrl, shortenedUrl)
	w.Write([]byte(fullShortenedUrl))
}

func (h *UrlHandler) ShortenUrl(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Unsupported content-type", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var request model.ApiShortenURLRequest
	err := decoder.Decode(&request)

	if err != nil {
		http.Error(w, "Failed to validate body", http.StatusBadRequest)
		return
	}

	if request.Url == "" {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	shortenedUrl, err := h.service.CreateShortUrl(r.Context(), request.Url)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(model.ApiShortenURLResponse{
		Result: fmt.Sprintf("%s/%s", h.cfg.BaseUrl, shortenedUrl),
	})
}
