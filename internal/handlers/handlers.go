package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener/cmd/config"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"

	"github.com/mi4r/go-url-shortener/internal/logger"
)

const (
	idLength = 8
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	URLMap = make(map[string]string)
	Flags  *config.Flags
)

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func ShortenURLHandler(w http.ResponseWriter, req *http.Request) {
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
		if _, exists := URLMap[shortID]; !exists {
			URLMap[shortID] = originalURL
			break
		}
	}

	shortURL := Flags.BaseShortAddr + "/" + shortID
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusBadRequest)
		logger.Sugar.Error("Failed to write response", zap.Error(err))
	}
}

func RedirectHandler(w http.ResponseWriter, req *http.Request) {
	shortID := chi.URLParam(req, "id")
	if len(shortID) == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	originalURL, exists := URLMap[shortID]

	if !exists {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, req, originalURL, http.StatusTemporaryRedirect)
}
