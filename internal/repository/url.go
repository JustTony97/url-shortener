package repository

type UrlRepository struct {
	cache map[string]string
}

func NewUrlRepository() *UrlRepository {
	return &UrlRepository{
		cache: make(map[string]string),
	}
}

func (r *UrlRepository) SetShortedUrl(shortedUrl string, originalUrl string) {
	r.cache[shortedUrl] = originalUrl
}

func (r *UrlRepository) GetOriginalUrl(shortedUrl string) (string, bool) {
	originalUrl, ok := r.cache[shortedUrl]
	return originalUrl, ok
}
