package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/handler"
	"github.com/JustTony97/url-shortener.git/internal/repository"
	"github.com/JustTony97/url-shortener.git/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	flag.Parse()
	r := chi.NewRouter()
	repo := repository.NewUrlRepository()
	service := service.NewUrlService(repo)
	handler := handler.NewUrlHandler(service)

	r.Post("/", handler.ShortUrl)
	r.Get("/{id}", handler.RedirectUrl)

	fmt.Printf("Starting listening on %s ...\n", config.RunAddr)
	err := http.ListenAndServe(config.RunAddr, r)
	if err != nil {
		panic(err)
	}
}
