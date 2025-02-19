package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mi4r/go-url-shortener/cmd/config"
	"github.com/mi4r/go-url-shortener/internal/auth"
	"github.com/mi4r/go-url-shortener/internal/logger"
	"github.com/mi4r/go-url-shortener/internal/storage"
	"github.com/mi4r/go-url-shortener/internal/storage/mocks"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortenURLHandler(t *testing.T) {
	mockStorage := new(mocks.MockStorage)
	mockStorage.On("Get", mock.Anything).Return(storage.URL{}, false)
	mockStorage.On("GetNextID").Return(1, nil)
	mockStorage.On("Save", mock.Anything).Return("", nil)
	mockStorage.On("Close").Return(nil)

	Flags = &config.Flags{
		BaseShortAddr: "http://short.url",
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("http://example.com"))
	w := httptest.NewRecorder()

	handler := ShortenURLHandler(mockStorage)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body := new(bytes.Buffer)
	body.ReadFrom(resp.Body)
	assert.Contains(t, body.String(), "http://short.url/")
	mockStorage.Close()
	mockStorage.AssertCalled(t, "Close")
}

func TestAPIShortenURLHandler(t *testing.T) {
	mockStorage := new(mocks.MockStorage)
	mockStorage.On("Get", mock.Anything).Return(storage.URL{}, false)
	mockStorage.On("GetNextID").Return(1, nil)
	mockStorage.On("Save", mock.Anything).Return("", nil)
	mockStorage.On("Close").Return(nil)

	Flags = &config.Flags{
		BaseShortAddr: "http://short.url",
	}

	reqBody := ShortenRequest{URL: "http://example.com"}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := APIShortenURLHandler(mockStorage)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var responseBody ShortenResponse
	json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.Contains(t, responseBody.Result, "http://short.url/")
	mockStorage.Close()
	mockStorage.AssertCalled(t, "Close")
}

func TestRedirectHandler(t *testing.T) {
	mockStorage := new(mocks.MockStorage)
	mockStorage.On("Get", "testID").Return(storage.URL{OriginalURL: "http://example.com"}, true)
	mockStorage.On("Close").Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/testID", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/{id}", RedirectHandler(mockStorage))
	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, "http://example.com", resp.Header.Get("Location"))
}

func TestPingHandler(t *testing.T) {
	logger.Sugar = *zap.NewNop().Sugar()
	mockStorage := new(mocks.MockStorage)
	mockStorage.On("Ping").Return(nil)
	mockStorage.On("Close").Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	handler := PingHandler(mockStorage)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUserURLsHandler(t *testing.T) {
	mockStorage := new(mocks.MockStorage)
	mockStorage.On("GetURLsByUserID", "userID").Return([]storage.URL{
		{ShortURL: "short1", OriginalURL: "http://example1.com"},
		{ShortURL: "short2", OriginalURL: "http://example2.com"},
	}, nil)

	Flags = &config.Flags{
		BaseShortAddr: "http://short.url",
	}

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	w := httptest.NewRecorder()

	auth.SetUserCookie(w, "userID")
	resp := w.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookie was set")
	}
	req.AddCookie(cookies[0])

	handler := UserURLsHandler(mockStorage)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBatchShortenURLHandler(t *testing.T) {
	mockStorage := new(mocks.MockStorage)
	mockStorage.On("SaveBatch", mock.Anything).Return([]string{"short1", "short2"}, nil)
	mockStorage.On("Close").Return(nil)

	Flags = &config.Flags{
		BaseShortAddr: "http://short.url",
	}

	reqBody := []BatchRequestItem{
		{CorrelationID: "1", OriginalURL: "http://example1.com"},
		{CorrelationID: "2", OriginalURL: "http://example2.com"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := BatchShortenURLHandler(mockStorage)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var responseBody []BatchResponseItem
	json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.Equal(t, 2, len(responseBody))
	assert.Equal(t, "http://short.url/short1", responseBody[0].ShortURL)
}

func TestDeleteUserURLsHandler(t *testing.T) {
	mockStorage := new(mocks.MockStorage)
	mockStorage.On("MarkURLsAsDeleted", "userID", []string{"id1", "id2"}).Return(nil)

	reqBody := []string{"id1", "id2"}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodDelete, "/user/urls", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	auth.SetUserCookie(w, "userID")
	resp := w.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()

	if len(cookies) == 0 {
		t.Fatal("No cookie was set")
	}
	req.AddCookie(cookies[0])

	handler := DeleteUserURLsHandler(mockStorage)
	handler.ServeHTTP(w, req)

	mockStorage.AssertCalled(t, "MarkURLsAsDeleted", "userID", []string{"id1", "id2"})
}
