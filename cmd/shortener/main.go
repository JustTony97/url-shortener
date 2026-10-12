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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	logger.InitLogger()
	var db *sqlx.DB

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	var repo repository.Repository

	if cfg.DatabaseDSN != "" {
		db, err = sqlx.Connect("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		logger.Log.Infof("Using db repository")

		driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			logger.Log.Fatalf("Failed to create migrate driver: %v", err)
		}

		m, err := migrate.NewWithDatabaseInstance(cfg.MigrationsPath, "postgres", driver)
		if err != nil {
			logger.Log.Fatalf("Failed to initialize migrations: %v", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Log.Fatalf("Failed to run migrations: %v", err)
		}
		defer m.Close()

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

	urlService := service.NewURLService(repo, cfg)
	urlHandler := handler.NewURLHandler(urlService, cfg)
	baseHandler := handler.NewBaseHandler(db)

	r.Post("/", middlewares.WithCompress(middlewares.WithDecompress(urlHandler.ShortURL)))
	r.Post("/api/shorten", middlewares.WithCompress(middlewares.WithDecompress(urlHandler.ShortenURL)))
	r.Post("/api/shorten/batch", middlewares.WithCompress(middlewares.WithDecompress(urlHandler.ShortenURLs)))
	r.Get("/{id}", urlHandler.RedirectURL)

	if db != nil {
		r.Get("/ping", baseHandler.Healthcheck)
	}

	log.Printf("Starting listening on %s ...\n", cfg.ServerAddr)
	err = http.ListenAndServe(cfg.ServerAddr, r)

	if err != nil {
		logger.Log.Fatal(err)
	}
}
