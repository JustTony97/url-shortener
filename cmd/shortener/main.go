package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
)

const appPort = ":8080"

var MemoryStorage map[string]string = map[string]string{}

func RedirectUrl(w http.ResponseWriter, r *http.Request) {
	shortedUrl := r.PathValue("id")
	if shortedUrl == "" {
		http.Error(w, "ID param is missing", http.StatusBadRequest)
		return
	}

	originalUrl, ok := MemoryStorage[shortedUrl]
	if !ok {
		http.Error(w, "Missing original URL", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
}

func ShortUrl(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
		http.Error(w, "Unsupported content-type", http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalUrl := string(bodyBytes)
	if originalUrl == "" {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	hash := fnv.New32a()
	hash.Write([]byte(originalUrl))
	shortedUrl := fmt.Sprintf("%x", hash.Sum32())

	MemoryStorage[shortedUrl] = originalUrl

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fullShortedURL := fmt.Sprintf("http://localhost:8080/%s", shortedUrl)
	w.Write([]byte(fullShortedURL))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", ShortUrl)
	mux.HandleFunc("GET /{id}", RedirectUrl)

	fmt.Printf("Starting listening on %s ...\n", appPort)
	err := http.ListenAndServe(appPort, mux)
	if err != nil {
		panic(err)
	}
}
