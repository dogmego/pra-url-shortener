package main

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"sync"
)

var (
	urlStore = make(map[string]string)
	mu       sync.RWMutex
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)

	if err := http.ListenAndServe(":8085", mux); err != nil {
		panic(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if r.URL.Path == "/" {
			handleShortenURL(w, r)
		} else {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	case http.MethodGet:
		handleRedirect(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	longURL := string(body)

	shortID := generateShortID(longURL)

	mu.Lock()
	urlStore[shortID] = longURL
	mu.Unlock()

	shortURL := "http://localhost:8085/" + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(shortURL))
}

func generateShortID(longURL string) string {
	hash := sha256.Sum256([]byte(longURL))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortID := r.URL.Path[1:]

	if shortID == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	mu.RLock()
	longURL, ok := urlStore[shortID]
	mu.RUnlock()

	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}
