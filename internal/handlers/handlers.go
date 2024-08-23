package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener/cmd/config"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"

	"github.com/mi4r/go-url-shortener/internal/logger"
	"github.com/mi4r/go-url-shortener/internal/storage"
)

const (
	idLength = 8
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	Flags *config.Flags
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func ShortenURLHandler(storageImpl storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
			if _, exists := storageImpl.Get(shortID); !exists {
				nextID, err := storageImpl.GetNextID()
				if err != nil {
					http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
					return
				}
				url := storage.URL{
					CorrelationID: strconv.Itoa(nextID),
					ShortURL:      shortID,
					OriginalURL:   originalURL,
				}
				existingURL, err := storageImpl.Save(url)
				if err != nil {
					http.Error(w, "Failed to save data", http.StatusInternalServerError)
					logger.Sugar.Error("Failed to save data:", zap.Error(err))
					return
				}
				if existingURL != "" {
					shortURL := Flags.BaseShortAddr + "/" + existingURL
					w.WriteHeader(http.StatusConflict)
					_, err = w.Write([]byte(shortURL))
					if err != nil {
						http.Error(w, "Failed to write response", http.StatusBadRequest)
						logger.Sugar.Error("Failed to write response", zap.Error(err))
					}
					return
				}
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
}

func APIShortenURLHandler(storageImpl storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}

		var requestBody ShortenRequest

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&requestBody); err != nil || requestBody.URL == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		originalURL := requestBody.URL

		var shortID string
		for {
			shortID = generateShortID()
			if _, exists := storageImpl.Get(shortID); !exists {
				nextID, err := storageImpl.GetNextID()
				if err != nil {
					http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
					return
				}
				url := storage.URL{
					CorrelationID: strconv.Itoa(nextID),
					ShortURL:      shortID,
					OriginalURL:   originalURL,
				}
				existingURL, err := storageImpl.Save(url)
				if err != nil {
					http.Error(w, "Failed to save data", http.StatusInternalServerError)
					logger.Sugar.Error("Failed to save data:", zap.Error(err))
					return
				}
				if existingURL != "" {
					shortURL := Flags.BaseShortAddr + "/" + existingURL

					responseBody := ShortenResponse{
						Result: shortURL,
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusConflict)
					if err := json.NewEncoder(w).Encode(responseBody); err != nil {
						http.Error(w, "Failed to write response", http.StatusBadRequest)
						logger.Sugar.Error("Failed to write response", zap.Error(err))
					}
					return
				}
				break
			}
		}

		shortURL := Flags.BaseShortAddr + "/" + shortID

		responseBody := ShortenResponse{
			Result: shortURL,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(responseBody); err != nil {
			http.Error(w, "Failed to write response", http.StatusBadRequest)
			logger.Sugar.Error("Failed to write response", zap.Error(err))
		}
	}
}

func BatchShortenURLHandler(storageImpl storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}

		var batchRequest []BatchRequestItem

		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		if err = json.Unmarshal(reqBody, &batchRequest); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if len(batchRequest) == 0 {
			http.Error(w, "Batch cannot be empty", http.StatusBadRequest)
			return
		}

		urls := make([]storage.URL, len(batchRequest))
		for i, item := range batchRequest {
			urls[i] = storage.URL{CorrelationID: item.CorrelationID, OriginalURL: item.OriginalURL}
		}

		shortIDs, err := storageImpl.SaveBatch(urls)
		if err != nil {
			http.Error(w, "Failed to save URL batch", http.StatusInternalServerError)
			logger.Sugar.Error("Failed to save URL batch: ", zap.Error(err))
			return
		}

		batchResponse := make([]BatchResponseItem, len(batchRequest))
		for i, shortID := range shortIDs {
			batchResponse[i] = BatchResponseItem{
				CorrelationID: batchRequest[i].CorrelationID,
				ShortURL:      Flags.BaseShortAddr + "/" + shortID,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(batchResponse); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			logger.Sugar.Error("Failed to encode response: ", zap.Error(err))
		}
	}
}

func RedirectHandler(storageImpl storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		shortID := chi.URLParam(req, "id")
		if len(shortID) == 0 {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		url, exists := storageImpl.Get(shortID)
		if !exists {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		http.Redirect(w, req, url.OriginalURL, http.StatusTemporaryRedirect)
	}
}

func PingHandler(storageImpl storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		pinger, ok := storageImpl.(storage.Pinger)
		if !ok {
			http.Error(w, "Connection to the database is not verified", http.StatusInternalServerError)
			logger.Sugar.Error("Connection to the database is not verified")
			return
		}
		if err := pinger.Ping(); err != nil {
			http.Error(w, "Connection to the database is not verified", http.StatusInternalServerError)
			logger.Sugar.Error("Connection to the database is not verified: ", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
