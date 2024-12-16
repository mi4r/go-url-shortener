package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenURL(t *testing.T) {
	// Моковый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		url := r.FormValue("url")
		if url != "https://example.com" {
			t.Errorf("Expected URL to be 'https://example.com', got %s", url)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"http://short.url/abc123"}`))
	}))
	defer ts.Close()

	client := &http.Client{}
	result, err := shortenURL(ts.URL, "https://example.com", client)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `{"result":"http://short.url/abc123"}`
	if result != expected {
		t.Errorf("Expected result %s, got %s", expected, result)
	}
}

func TestShortenURL_ErrorResponse(t *testing.T) {
	// Моковый сервер с ошибкой
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &http.Client{}
	_, err := shortenURL(ts.URL, "https://example.com", client)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
}

func TestShortenURL_InvalidURL(t *testing.T) {
	client := &http.Client{}
	_, err := shortenURL("invalid-url", "https://example.com", client)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
}
