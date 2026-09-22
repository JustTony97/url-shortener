package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/jmoiron/sqlx"
)

type DatabaseRepository struct {
	db *sqlx.DB
}

func NewDatabaseRepository(db *sqlx.DB) *DatabaseRepository {
	return &DatabaseRepository{
		db: db,
	}
}

func (r *DatabaseRepository) SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string) error {
	queryInsert := `
		INSERT INTO urls (short_url, original_url) 
		VALUES ($1, $2);`

	_, err := r.db.ExecContext(ctx, queryInsert, shortenedURL, originalURL)
	if err != nil {
		var pgErr interface {
			SQLState() string
		}

		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			querySelect := `SELECT short_url FROM urls WHERE original_url = $1;`
			logger.Log.Infoln("EXISTING ", querySelect)
			var existingShortURL string

			selectErr := r.db.QueryRowContext(ctx, querySelect, originalURL).Scan(&existingShortURL)
			if selectErr != nil {
				logger.Log.Errorf("failed to fetch existing url after conflict: %v", selectErr)
				return selectErr
			}

			return &model.ErrConflictWithExistingURL{ShortURL: existingShortURL}
		}

		logger.Log.Errorf("failed to insert url in database: %v", err)
		return err
	}

	return nil
}

func (r *DatabaseRepository) SetShortenedURLs(ctx context.Context, items []model.URL) error {
	query := `
		INSERT INTO urls (short_url, original_url) 
		VALUES (:short_url, :original_url)
		ON CONFLICT (short_url) 
		DO UPDATE SET original_url = EXCLUDED.original_url;`

	_, err := r.db.NamedExecContext(ctx, query, items)
	if err != nil {
		logger.Log.Errorf("failed to batch insert/update urls in database: %v", err)
		return err
	}
	return nil
}

func (r *DatabaseRepository) GetOriginalURL(ctx context.Context, shortenedURL string) (string, bool, error) {
	query := `SELECT original_url FROM urls WHERE short_url = $1 LIMIT 1;`

	var originalURL string
	err := r.db.GetContext(ctx, &originalURL, query, shortenedURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}

		logger.Log.Errorf("failed to get original url from database: %v", err)
		return "", false, err
	}

	return originalURL, true, nil
}
