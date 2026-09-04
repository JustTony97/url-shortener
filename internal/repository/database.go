package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JustTony97/url-shortener.git/internal/logger"
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
	query := `
		INSERT INTO urls (short_url, original_url) 
		VALUES ($1, $2)
		ON CONFLICT (short_url) 
		DO UPDATE SET original_url = EXCLUDED.original_url;`

	_, err := r.db.ExecContext(ctx, query, shortenedURL, originalURL)
	if err != nil {
		logger.Log.Errorf("failed to insert/update url in database: %v", err)
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
