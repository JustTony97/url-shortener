package main

import (
	"fmt"
	"net/http"

	"github.com/JustTony97/url-shortener.git/internal/handler"
	"github.com/JustTony97/url-shortener.git/internal/repository"
	"github.com/JustTony97/url-shortener.git/internal/service"
	"github.com/go-chi/chi"
)

const appPort = ":8080"

func main() {
	r := chi.NewRouter()
	repo := repository.NewUrlRepository()
	service := service.NewUrlService(repo)
	handler := handler.NewUrlHandler(service)

	r.Post("/", handler.ShortUrl)
	r.Get("/{id}", handler.RedirectUrl)

	fmt.Printf("Starting listening on %s ...\n", appPort)
	err := http.ListenAndServe(appPort, r)
	if err != nil {
		panic(err)
	}
}
