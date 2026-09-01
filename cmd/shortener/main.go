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
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/go-chi/chi/v5"
)

func main() {
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db := sqlx.MustOpen("pgx", cfg.DatabaseDSN)
	defer db.Close()
	// err = db.Ping()
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)
	repo := repository.NewUrlRepository(cfg)
	err = repo.Load()
	if err != nil {
		log.Fatal(err)
	}

	service := service.NewUrlService(repo)
	urlHandler := handler.NewUrlHandler(service, cfg)
	baseHandler := handler.NewBaseHandler(db)

	r.Post("/", middlewares.WithCompress(middlewares.WithDecompress(urlHandler.ShortUrl)))
	r.Post("/api/shorten", middlewares.WithCompress(middlewares.WithDecompress(urlHandler.ShortenUrl)))
	r.Get("/{id}", urlHandler.RedirectUrl)
	r.Get("/ping", baseHandler.Healthcheck)

	log.Printf("Starting listening on %s ...\n", cfg.ServerAddr)
	err = http.ListenAndServe(cfg.ServerAddr, r)

	if err != nil {
		log.Fatal(err)
	}
}
