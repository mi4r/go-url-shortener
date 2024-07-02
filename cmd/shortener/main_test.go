package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		existedURLMap map[string]string
		shorten       string
		method        string
		want          want
	}{
		{
			name:          "success case",
			method:        http.MethodGet,
			shorten:       "/abc",
			existedURLMap: map[string]string{"abc": "http://example.com"},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				origin:     "http://example.com",
			},
		},
		{
			name:          "invalid short ID",
			method:        http.MethodGet,
			shorten:       "/invalid",
			existedURLMap: map[string]string{"abc": "http://example.com"},
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
			urlMap = tt.existedURLMap
			req := httptest.NewRequest(tt.method, tt.shorten, nil)
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(redirectHandler)
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
	urlMap = make(map[string]string)
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
			handler := http.HandlerFunc(shortenURLHandler)
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
