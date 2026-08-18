package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/handler"
	"github.com/JustTony97/url-shortener.git/internal/repository"
	"github.com/JustTony97/url-shortener.git/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	repo := repository.NewUrlRepository()
	service := service.NewUrlService(repo)
	handler := handler.NewUrlHandler(service, cfg)

	r.Post("/", handler.ShortUrl)
	r.Get("/{id}", handler.RedirectUrl)

	fmt.Printf("Starting listening on %s ...\n", cfg.ServerAddr)
	err = http.ListenAndServe(cfg.ServerAddr, r)

	if err != nil {
		log.Fatal(err)
	}
}
