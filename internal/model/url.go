package model

type URL struct {
	UUID        string `json:"uuid" db:"uuid"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
	UserID      string `json:"user_id" db:"user_id"`
	IsDeleted   bool   `db:"is_deleted"`
}

type DeleteURLTask struct {
	UserID    string
	ShortURLs []string
}
