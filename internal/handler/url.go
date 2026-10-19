package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/middlewares"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/go-chi/chi/v5"

	"net/http"
	"strings"
)

type URLHandler struct {
	service URLService
	cfg     config.Config
}

type URLService interface {
	GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error)
	CreateShortURL(ctx context.Context, originalURL string, userID string) (string, error)
	CreateShortURLs(ctx context.Context, reqItems []model.BatchItemRequest, userID string) ([]model.BatchResponseItem, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.URL, error)
}

func NewURLHandler(s URLService, c config.Config) *URLHandler {
	return &URLHandler{service: s, cfg: c}
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortenedURL := chi.URLParam(r, "id")
	if shortenedURL == "" {
		http.Error(w, "ID param is missing", http.StatusBadRequest)
		return
	}

	originalURL, ok, err := h.service.GetOriginalURL(r.Context(), shortenedURL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Missing original URL", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *URLHandler) ShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middlewares.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
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

	originalURL := string(bodyBytes)
	if originalURL == "" {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.service.CreateShortURL(r.Context(), originalURL, userID)
	if err != nil {
		var conflictErr *model.ErrConflictWithExistingURL
		if errors.As(err, &conflictErr) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf("%s/%s", h.cfg.BaseURL, conflictErr.ShortURL)))
			return
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fullShortenedURL := fmt.Sprintf("%s/%s", h.cfg.BaseURL, shortenedURL)
	w.Write([]byte(fullShortenedURL))
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middlewares.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Unsupported content-type", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var request model.APIShortenURLRequest
	err := decoder.Decode(&request)

	if err != nil {
		http.Error(w, "Failed to validate body", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.service.CreateShortURL(r.Context(), request.URL, userID)

	if err != nil {
		var conflictErr *model.ErrConflictWithExistingURL
		if errors.As(err, &conflictErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)

			encoder := json.NewEncoder(w)
			encoder.Encode(model.APIShortenURLResponse{
				Result: fmt.Sprintf("%s/%s", h.cfg.BaseURL, conflictErr.ShortURL),
			})
			return
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(model.APIShortenURLResponse{
		Result: fmt.Sprintf("%s/%s", h.cfg.BaseURL, shortenedURL),
	})
}

func (h *URLHandler) ShortenURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middlewares.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Unsupported content-type", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var request []model.BatchItemRequest
	err := decoder.Decode(&request)

	if err != nil {
		http.Error(w, "Failed to validate body", http.StatusBadRequest)
		return
	}

	if len(request) == 0 {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	savedItems, err := h.service.CreateShortURLs(r.Context(), request, userID)

	if err != nil {
		if errors.Is(err, model.ErrConflictURL) {
			http.Error(w, "One or more URLs already exist", http.StatusConflict)
			return
		}
		logger.Log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(savedItems)
}

func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middlewares.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.service.GetUserURLs(ctx, userID)

	if err != nil {
		logger.Log.Errorf("failed to get user urls: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make(model.APIUserShortenBatchResponse, 0, len(urls))
	for _, url := range urls {
		response = append(response, model.UserBatchResponseItem{
			OriginalURL: url.OriginalURL,
			ShortURL:    fmt.Sprintf("%s/%s", h.cfg.BaseURL, url.ShortURL),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}
