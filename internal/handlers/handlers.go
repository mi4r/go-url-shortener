package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener/cmd/config"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"

	"github.com/mi4r/go-url-shortener/internal/auth"
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

type URLResponseItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

		userID, valid := auth.ValidateUserCookie(req)
		if !valid {
			userID = auth.GenerateUserID()
			auth.SetUserCookie(w, userID)
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
					UserID:        userID,
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

		userID, valid := auth.ValidateUserCookie(req)
		if !valid {
			userID = auth.GenerateUserID()
			auth.SetUserCookie(w, userID)
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
					UserID:        userID,
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

		userID, valid := auth.ValidateUserCookie(req)
		if !valid {
			userID = auth.GenerateUserID()
			auth.SetUserCookie(w, userID)
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
			urls[i] = storage.URL{CorrelationID: item.CorrelationID, OriginalURL: item.OriginalURL, UserID: userID}
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

		if url.DeletedFlag {
			http.Error(w, "Gone", http.StatusGone)
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

// UserURLsHandler возвращает все URL, сокращенные текущим пользователем.
func UserURLsHandler(storageImpl storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Проверяем подлинность куки
		userID, valid := auth.ValidateUserCookie(req)
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Получаем URL'ы пользователя из хранилища
		urls, err := storageImpl.GetURLsByUserID(userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Проверяем, есть ли сокращенные URL'ы
		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Формируем ответ
		response := make([]URLResponseItem, len(urls))
		for i, url := range urls {
			response[i] = URLResponseItem{
				ShortURL:    Flags.BaseShortAddr + "/" + url.ShortURL,
				OriginalURL: url.OriginalURL,
			}
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func DeleteUserURLsHandler(storageImpl storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, valid := auth.ValidateUserCookie(req)
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		logger.Sugar.Infof("userID: %v", userID)

		var ids []string
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&ids); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		idChan := make(chan string)
		// done := make(chan struct{})
		var wg sync.WaitGroup

		go func() {
			defer close(idChan)
			for _, id := range ids {
				idChan <- id
			}
		}()

		updateBatch := func(urls []string) {
			if len(urls) == 0 {
				return
			}
			logger.Sugar.Infof("urls: %v", urls)
			if err := storageImpl.MarkURLsAsDeleted(userID, urls); err != nil {
				logger.Sugar.Errorf("Error marking URLs as deleted: %v", err)
			}
		}

		numWorkers := 5
		batchSize := 10
		urlsBatch := make([]string, 0, batchSize)
		urlsChan := make(chan []string, len(urlsBatch))

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(inputCh chan []string, wg *sync.WaitGroup) {
				defer wg.Done()
				for batch := range inputCh {
					updateBatch(batch)
				}
			}(urlsChan, &wg)
		}

		go func() {
			defer close(urlsChan)
			for id := range idChan {
				urlsBatch = append(urlsBatch, id)
				if len(urlsBatch) >= batchSize {
					urlsChan <- urlsBatch
					urlsBatch = make([]string, 0, batchSize)
				}
			}
			if len(urlsBatch) > 0 {
				urlsChan <- urlsBatch
			}
			// done <- struct{}{}
		}()

		// <-done
		// t := time.NewTimer(time.Second * 10)
		// select {
		// case <-done:
		// 	logger.Sugar.Info("Получен done")
		// case <-t.C:
		// 	logger.Sugar.Info("timer")
		// }
		wg.Wait()
		w.WriteHeader(http.StatusAccepted)
	}
}
