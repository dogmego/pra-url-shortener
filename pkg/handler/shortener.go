package handler

import (
	"io"
	"net/http"
	"sync"
)

var (
	urlStore = make(map[string]string)
	mu       sync.RWMutex
)

func HandleRedirect(w http.ResponseWriter, r *http.Request) {
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

func HandleShortenURL(w http.ResponseWriter, r *http.Request) {
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

	shortID := GenerateShortID(longURL)

	mu.Lock()
	urlStore[shortID] = longURL
	mu.Unlock()

	shortURL := "http://localhost:8085/" + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(shortURL))
}
