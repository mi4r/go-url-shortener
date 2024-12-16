package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestLoggingResponseWriter_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	responseData := &responseData{}
	lrw := loggingResponseWriter{
		ResponseWriter: rr,
		responseData:   responseData,
	}

	data := []byte("test response")
	size, err := lrw.Write(data)
	if err != nil {
		t.Fatalf("Unexpected error writing data: %v", err)
	}

	if size != len(data) {
		t.Errorf("Expected written size %d, got %d", len(data), size)
	}

	if responseData.size != len(data) {
		t.Errorf("Expected responseData.size %d, got %d", len(data), responseData.size)
	}
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	responseData := &responseData{}
	lrw := loggingResponseWriter{
		ResponseWriter: rr,
		responseData:   responseData,
	}

	status := http.StatusCreated
	lrw.WriteHeader(status)

	if responseData.status != status {
		t.Errorf("Expected responseData.status %d, got %d", status, responseData.status)
	}

	if rr.Code != status {
		t.Errorf("Expected status code %d, got %d", status, rr.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Замена глобального логгера на безопасный для тестов
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	Sugar = *logger.Sugar()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware := LoggingMiddleware(handler)
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "Hello, world!" {
		t.Errorf("Expected response body 'Hello, world!', got '%s'", rr.Body.String())
	}
}

func TestLoggingMiddleware_CaptureResponseData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	Sugar = *logger.Sugar()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rr := httptest.NewRecorder()

	middleware := LoggingMiddleware(handler)
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rr.Code)
	}

	if rr.Body.String() != "Not Found" {
		t.Errorf("Expected response body 'Not Found', got '%s'", rr.Body.String())
	}
}
