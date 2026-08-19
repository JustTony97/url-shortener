package main

import (
	"log"
	"net/http"

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
	repo := repository.NewUrlRepository()
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
