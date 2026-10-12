package model

import (
	"errors"
	"fmt"
)

var (
	ErrConflictURL   = errors.New("url already exists")
	ErrHashCollision = errors.New("hash collision")
)

type ErrConflictWithExistingURL struct {
	ShortURL string
}

func (e *ErrConflictWithExistingURL) Error() string {
	return fmt.Sprintf("original url already exists in database with short url: %s", e.ShortURL)
}

func (e *ErrConflictWithExistingURL) Unwrap() error {
	return ErrConflictURL
}
