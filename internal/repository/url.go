package repository

type UrlRepository map[string]string

func NewUrlRepository() UrlRepository {
	return make(UrlRepository)
}

func (r UrlRepository) SetShortedUrl(shortedUrl string, originalUrl string) {
	r[shortedUrl] = originalUrl
}

func (r UrlRepository) GetOriginalUrl(shortedUrl string) (string, bool) {
	originalUrl, ok := r[shortedUrl]
	return originalUrl, ok
}
