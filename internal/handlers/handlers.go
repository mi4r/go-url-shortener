package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

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
	URLMap   = make(map[string]URL)
	Flags    *config.Flags
	Database *sql.DB
)

type URL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func SaveToFile(filePath string, url URL) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Sugar.Error(err)
		}
	}()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(url); err != nil {
		return err
	}
	return nil
}

func LoadFromFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Sugar.Error(err)
		}
	}()

	decoder := json.NewDecoder(file)
	for {
		var url URL
		if err := decoder.Decode(&url); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		URLMap[url.ShortURL] = url
	}
	return nil
}

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
			url := URL{
				UUID:        strconv.Itoa(len(URLMap) + 1),
				ShortURL:    shortID,
				OriginalURL: originalURL,
			}
			URLMap[shortID] = url

			if err := SaveToFile(Flags.URLStorageFilePath, url); err != nil {
				http.Error(w, "Failed to save data", http.StatusInternalServerError)
				logger.Sugar.Error("Failed to save data:", zap.Error(err))
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

func APIShortenURLHandler(w http.ResponseWriter, req *http.Request) {
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
		if _, exists := URLMap[shortID]; !exists {
			url := URL{
				UUID:        strconv.Itoa(len(URLMap) + 1),
				ShortURL:    shortID,
				OriginalURL: originalURL,
			}
			URLMap[shortID] = url

			if err := SaveToFile(Flags.URLStorageFilePath, url); err != nil {
				http.Error(w, "Failed to save data", http.StatusInternalServerError)
				logger.Sugar.Error("Failed to save data:", zap.Error(err))
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

func RedirectHandler(w http.ResponseWriter, req *http.Request) {
	shortID := chi.URLParam(req, "id")
	if len(shortID) == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	url, exists := URLMap[shortID]

	if !exists {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, req, url.OriginalURL, http.StatusTemporaryRedirect)
}

func PingHandler(w http.ResponseWriter, req *http.Request) {
	if err := Database.Ping(); err != nil {
		http.Error(w, "Connection to the database is not verified", http.StatusInternalServerError)
		logger.Sugar.Error("Connection to the database is not verified: ", zap.Error(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
