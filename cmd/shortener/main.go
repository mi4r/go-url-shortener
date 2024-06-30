package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

const (
	addr     = "localhost:8080"
	idLength = 8
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	urlMap = make(map[string]string)
)

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	originalURL := string(body)

	var shortID string
	for {
		shortID = generateShortID()
		if _, exists := urlMap[shortID]; !exists {
			urlMap[shortID] = originalURL
			break
		}
	}

	shortURL := "http://localhost:8080/" + shortID
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := strings.TrimPrefix(r.URL.Path, "/")
	fmt.Println(shortID)
	if len(shortID) == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	originalURL, exists := urlMap[shortID]

	if !exists {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	fmt.Println(originalURL)
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortenURLHandler)
	mux.HandleFunc("/{id}", redirectHandler)
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}
