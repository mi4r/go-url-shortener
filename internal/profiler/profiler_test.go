package profiler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNoCacheHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})
	noCacheHandler := NoCache(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-Modified-Since", "Wed, 21 Oct 2015 07:28:00 GMT")
	req.Header.Set("ETag", "some-etag-value")

	rr := httptest.NewRecorder()
	noCacheHandler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	// Проверяем заголовки отключения кэширования
	for k, v := range noCacheHeaders {
		if got := resp.Header.Get(k); got != v {
			t.Errorf("Header %q: expected %q, got %q", k, v, got)
		}
	}

	// Проверяем, что заголовки ETag удалены
	for _, header := range etagHeaders {
		if _, exists := req.Header[header]; exists {
			t.Errorf("Header %q was not removed from the request", header)
		}
	}
}

func TestNoCachePassThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})
	noCacheHandler := NoCache(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	noCacheHandler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Проверяем тело ответа
	body := rr.Body.String()
	if !strings.Contains(body, "test response") {
		t.Errorf("Expected response body to contain %q, got %q", "test response", body)
	}
}

func TestProfilerRoutes(t *testing.T) {
	profilerHandler := Profiler()

	tests := []struct {
		path       string
		statusCode int
	}{
		{path: "/", statusCode: http.StatusMovedPermanently},
		{path: "/pprof", statusCode: http.StatusMovedPermanently},
		{path: "/pprof/", statusCode: http.StatusOK},
		{path: "/pprof/cmdline", statusCode: http.StatusOK},
		{path: "/pprof/profile", statusCode: http.StatusOK},
		{path: "/pprof/symbol", statusCode: http.StatusOK},
		{path: "/pprof/trace", statusCode: http.StatusOK},
		{path: "/pprof/goroutine", statusCode: http.StatusOK},
		{path: "/pprof/threadcreate", statusCode: http.StatusOK},
		{path: "/pprof/mutex", statusCode: http.StatusOK},
		{path: "/pprof/heap", statusCode: http.StatusOK},
		{path: "/pprof/block", statusCode: http.StatusOK},
		{path: "/pprof/allocs", statusCode: http.StatusOK},
		{path: "/vars", statusCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			profilerHandler.ServeHTTP(rec, req)

			if rec.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d for path %q", tt.statusCode, rec.Code, tt.path)
			}
		})
	}
}
