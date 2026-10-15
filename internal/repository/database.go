package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

func (r *DatabaseRepository) GetOriginalURL(ctx context.Context, shortenedURL string) (*model.URL, error) {
	query := `SELECT short_url, is_deleted FROM urls WHERE short_url = $1 LIMIT 1;`

	var urlData model.URL
	err := r.db.GetContext(ctx, &urlData, query, shortenedURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: failed to get url state: %w", err)
	}

	return &urlData, nil
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

func (r *DatabaseRepository) DeleteUserURLs(ctx context.Context, userIDs []string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	query := `
		UPDATE urls 
		SET is_deleted = true 
		FROM unnest($1::text[], $2::text[]) AS batch(user_id, short_url)
		WHERE urls.user_id = batch.user_id AND urls.short_url = batch.short_url;
	`
	_, err := r.db.ExecContext(ctx, query, pq.Array(userIDs), pq.Array(shortURLs))

	if err != nil {
		logger.Log.Error(err)
		return err
	}

	return nil
}
