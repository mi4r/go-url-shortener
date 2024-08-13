package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener/cmd/config"
	"github.com/mi4r/go-url-shortener/internal/handlers"
	"github.com/mi4r/go-url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	type want struct {
		statusCode int
		origin     string
	}
	tests := []struct {
		name          string
		existedURLMap map[string]storage.URL
		shorten       string
		method        string
		want          want
	}{
		{
			name:          "success case",
			method:        http.MethodGet,
			shorten:       "/abc",
			existedURLMap: map[string]storage.URL{"abc": {ShortURL: "abc", OriginalURL: "http://example.com"}},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				origin:     "http://example.com",
			},
		},
		{
			name:          "invalid short ID",
			method:        http.MethodGet,
			shorten:       "/invalid",
			existedURLMap: map[string]storage.URL{"abc": {ShortURL: "abc", OriginalURL: "http://example.com"}},
			want: want{
				statusCode: http.StatusBadRequest,
				origin:     "",
			},
		},
		{
			name:          "no short ID",
			method:        http.MethodGet,
			shorten:       "/",
			existedURLMap: nil,
			want: want{
				statusCode: http.StatusBadRequest,
				origin:     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := storage.NewMemoryStorage()
			for _, url := range tt.existedURLMap {
				_, _ = storage.Save(url)
			}
			req := httptest.NewRequest(tt.method, tt.shorten, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", strings.TrimPrefix(tt.shorten, "/"))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(handlers.RedirectHandler(storage))
			handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			if tt.want.statusCode == http.StatusTemporaryRedirect {
				location := w.Header().Get("Location")
				assert.Equal(t, tt.want.origin, location)
			} else {
				respBody := w.Body.String()
				assert.Contains(t, respBody, "Invalid request")
			}
		})
	}
}

func TestShortenURLHandler(t *testing.T) {
	storage := storage.NewMemoryStorage()
	handlers.Flags = &config.Flags{
		RunAddr:            "localhost:8080",
		BaseShortAddr:      "http://localhost:8080",
		URLStorageFilePath: "test_storage.json",
		DataBaseDSN:        "host=localhost user=url_storage password=1234 dbname=url_storage sslmode=disable",
	}
	defer os.Remove("test_storage.json")

	type want struct {
		expectedCode int
		expectedResp string
	}
	tests := []struct {
		name        string
		method      string
		originalURL string
		want        want
	}{
		{
			name:        "sucess case",
			method:      http.MethodPost,
			originalURL: "http://example.com",
			want: want{
				expectedCode: http.StatusCreated,
				expectedResp: "http://localhost:8080/",
			},
		},
		{
			name:        "empty case",
			method:      http.MethodPost,
			originalURL: "",
			want: want{
				expectedCode: http.StatusBadRequest,
				expectedResp: "Invalid request body\n",
			},
		},
		{
			name:        "invalid method case",
			method:      http.MethodGet,
			originalURL: "http://example.com",
			want: want{
				expectedCode: http.StatusBadRequest,
				expectedResp: "Invalid request method\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.originalURL))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(handlers.ShortenURLHandler(storage))
			handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.expectedCode, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.want.expectedCode == http.StatusCreated {
				assert.True(t, strings.HasPrefix(string(resBody), tt.want.expectedResp))
			} else {
				assert.Equal(t, tt.want.expectedResp, string(resBody))
			}
		})
	}
}

func TestAPIShortenURLHandler(t *testing.T) {
	storage := storage.NewMemoryStorage()
	handlers.Flags = &config.Flags{
		RunAddr:            "localhost:8080",
		BaseShortAddr:      "http://localhost:8080",
		URLStorageFilePath: "test_storage.json",
		DataBaseDSN:        "host=localhost user=url_storage password=1234 dbname=url_storage sslmode=disable",
	}
	defer os.Remove("test_storage.json")

	type want struct {
		expectedCode int
		expectedResp string
	}
	tests := []struct {
		name        string
		method      string
		requestBody map[string]string
		want        want
	}{
		{
			name:        "sucess case",
			method:      http.MethodPost,
			requestBody: map[string]string{"url": "http://example.com"},
			want: want{
				expectedCode: http.StatusCreated,
				expectedResp: "http://localhost:8080/",
			},
		},
		{
			name:        "empty case",
			method:      http.MethodPost,
			requestBody: map[string]string{},
			want: want{
				expectedCode: http.StatusBadRequest,
				expectedResp: "Invalid request body\n",
			},
		},
		{
			name:        "invalid method case",
			method:      http.MethodGet,
			requestBody: nil,
			want: want{
				expectedCode: http.StatusBadRequest,
				expectedResp: "Invalid request method\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if tt.requestBody != nil {
				reqBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			} else {
				reqBody = []byte("invalid json")
			}

			req := httptest.NewRequest(tt.method, "/api/shorten", bytes.NewReader(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler := http.HandlerFunc(handlers.APIShortenURLHandler(storage))
			handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.expectedCode, res.StatusCode)

			if tt.want.expectedCode == http.StatusCreated {
				var responseBody map[string]string

				if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}
				assert.Contains(t, responseBody["result"], "http://localhost:8080/", "Expected response to contain short URL")
			}
		})
	}
}
