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
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db := sqlx.MustOpen("pgx", cfg.DatabaseDSN)
	defer db.Close()

	var repo repository.Repository

	if cfg.DatabaseDSN != "" {
		logger.Log.Infof("Using db repository")

		driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			logger.Log.Fatalf("Failed to create migrate driver: %v", err)
		}

		m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
		if err != nil {
			logger.Log.Fatalf("Failed to initialize migrations: %v", err)
		}
		defer m.Close()

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Log.Fatalf("Failed to run migrations: %v", err)
		}
		repo = repository.NewDatabaseRepository(db)

	} else if cfg.FileStoragePath != "" {
		logger.Log.Infof("Using file repository")
		repo = repository.NewFileRepository(cfg)
	} else {
		logger.Log.Info("Using memory repository")
		repo = repository.NewMemoryRepository()
	}

	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)

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
