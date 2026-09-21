package model

import "errors"

var ErrConflictURL = errors.New("url already exists")

type ErrConflictWithExistingURL struct {
	ShortURL string
}

func (e *ErrConflictWithExistingURL) Error() string {
	return "original url already exists in database"
}
