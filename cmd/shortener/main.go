package main

import (
	"log"
	"net/http"
	"time"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/handler"
	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/middlewares"
	"github.com/JustTony97/url-shortener.git/internal/repository"
	"github.com/JustTony97/url-shortener.git/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)
	repo := repository.NewUrlRepository(cfg)
	err = repo.Load()
	if err != nil {
		log.Fatal(err)
	}
	go startPeriodicFlush(repo, 10*time.Second)

	service := service.NewUrlService(repo)
	handler := handler.NewUrlHandler(service, cfg)

	r.Post("/", middlewares.WithCompress(middlewares.WithDecompress(handler.ShortUrl)))
	r.Post("/api/shorten", middlewares.WithCompress(middlewares.WithDecompress(handler.ShortenUrl)))
	r.Get("/{id}", handler.RedirectUrl)

	log.Printf("Starting listening on %s ...\n", cfg.ServerAddr)
	err = http.ListenAndServe(cfg.ServerAddr, r)

	if err != nil {
		log.Fatal(err)
	}
}

func startPeriodicFlush(repo *repository.UrlRepository, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := repo.Flush(); err != nil {
			logger.Log.Errorf("Periodic flush failed: %v", err)
		}
	}
}
