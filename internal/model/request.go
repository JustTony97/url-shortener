package model

type APIShortenURLRequest struct {
	URL string `json:"url"`
}

type BatchItemRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type APIShortenURLsRequest []BatchItemRequest
