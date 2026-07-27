package main

import (
	"fmt"
	"net/http"

	"github.com/JustTony97/url-shortener.git/internal/handler"
	"github.com/JustTony97/url-shortener.git/internal/repository"
	"github.com/JustTony97/url-shortener.git/internal/service"
)

const appPort = ":8080"

func main() {
	mux := http.NewServeMux()
	repo := repository.NewUrlRepository()
	service := service.NewUrlService(repo)
	handler := handler.NewUrlHandler(service)

	mux.HandleFunc("POST /", handler.ShortUrl)
	mux.HandleFunc("GET /{id}", handler.RedirectUrl)

	fmt.Printf("Starting listening on %s ...\n", appPort)
	err := http.ListenAndServe(appPort, mux)
	if err != nil {
		panic(err)
	}
}
