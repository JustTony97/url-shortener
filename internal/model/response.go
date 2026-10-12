package model

type APIShortenURLResponse struct {
	Result string `json:"result"`
}

type BatchResponseItem struct {
	OriginalURL string `json:"correlation_id"`
	ShortURL    string `json:"short_url"`
}

type APIShortenBatchResponse []BatchResponseItem
