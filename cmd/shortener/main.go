package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener.git/cmd/config"
)

const (
	idLength = 8
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	urlMap = make(map[string]string)
	flags  *config.Flags
)

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func shortenURLHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
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

	shortURL := flags.BaseShortAddr + "/" + shortID
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusBadRequest)
		log.Println("Failed to write response:", err)
	}
}

func redirectHandler(w http.ResponseWriter, req *http.Request) {
	shortID := chi.URLParam(req, "id")
	if len(shortID) == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	originalURL, exists := urlMap[shortID]

	if !exists {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, req, originalURL, http.StatusTemporaryRedirect)
}

func main() {
	flags = config.Init()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", shortenURLHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", redirectHandler)
		})
	})

	log.Printf("Starting server on %s\n", flags.RunAddr)
	log.Fatal(http.ListenAndServe(flags.RunAddr, r))
}
