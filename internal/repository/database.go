package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	shortURLConstraint    = "idx_urls_short_url"
	originalURLConstraint = "idx_urls_original_url"
)

type DatabaseRepository struct {
	db *sqlx.DB
}

func NewDatabaseRepository(db *sqlx.DB) *DatabaseRepository {
	return &DatabaseRepository{
		db: db,
	}
}

func (r *DatabaseRepository) SetShortenedURL(ctx context.Context, shortenedURL string, originalURL string, userID string) error {
	query := `
		INSERT INTO urls (short_url, original_url, user_id) 
		VALUES ($1, $2, $3);`

	_, err := r.db.ExecContext(ctx, query, shortenedURL, originalURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case originalURLConstraint:
				return model.ErrHashCollision
			case shortURLConstraint:
				return model.ErrConflictURL
			}
		}
		logger.Log.Errorf("failed to insert url in database: %v", err)
		return err
	}

	return nil
}

func (r *DatabaseRepository) SetShortenedURLs(ctx context.Context, items []model.URL, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3);`

	for _, item := range items {
		_, err := tx.ExecContext(ctx, query, item.ShortURL, item.OriginalURL, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case originalURLConstraint:
					return model.ErrHashCollision
				case shortURLConstraint:
					return model.ErrConflictURL
				default:
					logger.Log.Errorf("unexpected constraint violation: %v", err)
					return err
				}
			}
			logger.Log.Errorf("failed to batch insert url: %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
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

func (r *DatabaseRepository) GetUserURLs(ctx context.Context, userID string) ([]model.URL, error) {
	var urls []model.URL

	query := `SELECT uuid, original_url, short_url FROM urls WHERE user_id = $1;`

	err := r.db.SelectContext(ctx, &urls, query, userID)

	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	return urls, nil
}
