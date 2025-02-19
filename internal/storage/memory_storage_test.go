package storage

import (
	"strconv"
	"testing"

	"github.com/mi4r/go-url-shortener/internal/logger"
	"go.uber.org/zap"
)

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	storage := NewMemoryStorage()

	url := URL{
		ShortURL:    "short123",
		OriginalURL: "https://example.com",
		UserID:      "user1",
	}

	_, err := storage.Save(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrievedURL, exists := storage.Get("short123")
	if !exists {
		t.Fatalf("URL not found")
	}

	if retrievedURL.OriginalURL != url.OriginalURL {
		t.Errorf("expected %s, got %s", url.OriginalURL, retrievedURL.OriginalURL)
	}
}

func TestMemoryStorage_SaveBatch(t *testing.T) {
	storage := NewMemoryStorage()

	urls := []URL{
		{OriginalURL: "https://example1.com", UserID: "user1"},
		{OriginalURL: "https://example2.com", UserID: "user1"},
	}

	ids, err := storage.SaveBatch(urls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ids) != len(urls) {
		t.Errorf("expected %d IDs, got %d", len(urls), len(ids))
	}

	for _, id := range ids {
		if _, exists := storage.Get(id); !exists {
			t.Errorf("URL with ID %s not found", id)
		}
	}
}

func TestMemoryStorage_GetURLsByUserID(t *testing.T) {
	storage := NewMemoryStorage()

	urls := []URL{
		{ShortURL: "short1", OriginalURL: "https://example1.com", UserID: "user1"},
		{ShortURL: "short2", OriginalURL: "https://example2.com", UserID: "user1"},
		{ShortURL: "short3", OriginalURL: "https://example3.com", UserID: "user2"},
	}

	for _, url := range urls {
		_, _ = storage.Save(url)
	}

	user1URLs, err := storage.GetURLsByUserID("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(user1URLs) != 2 {
		t.Errorf("expected 2 URLs for user1, got %d", len(user1URLs))
	}

	user2URLs, err := storage.GetURLsByUserID("user2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(user2URLs) != 1 {
		t.Errorf("expected 1 URL for user2, got %d", len(user2URLs))
	}
}

func TestMemoryStorage_GetNextID(t *testing.T) {
	storage := NewMemoryStorage()

	id, err := storage.GetNextID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 1 {
		t.Errorf("expected next ID to be 1, got %d", id)
	}

	// Save a URL to increment the ID.
	_, _ = storage.Save(URL{ShortURL: "short1", OriginalURL: "https://example.com", UserID: "user1"})

	id, err = storage.GetNextID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 2 {
		t.Errorf("expected next ID to be 2, got %d", id)
	}
}

func TestMemoryStorage_MarkURLsAsDeleted(t *testing.T) {
	logger.Sugar = *zap.NewNop().Sugar()
	storage := NewMemoryStorage()

	url := URL{
		ShortURL:    "short123",
		OriginalURL: "https://example.com",
		UserID:      "user1",
	}

	_, _ = storage.Save(url)

	err := storage.MarkURLsAsDeleted("user1", []string{"short123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrievedURL, exists := storage.Get("short123")
	if !exists {
		t.Fatalf("URL not found")
	}

	if !retrievedURL.DeletedFlag {
		t.Errorf("expected DeletedFlag to be true, got false")
	}
}

func TestMemoryStorage_Close(t *testing.T) {
	s := NewMemoryStorage()
	err := s.Close()
	if err != nil {
		t.Errorf("close of storage is failed")
	}
}

func TestURLCount(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int
		want     int
	}{
		{
			name:     "empty storage",
			dataSize: 0,
			want:     0,
		},
		{
			name:     "single URL",
			dataSize: 1,
			want:     1,
		},
		{
			name:     "multiple URLs",
			dataSize: 5,
			want:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём хранилище с тестовыми данными
			storage := &MemoryStorage{
				data: make(map[string]URL, tt.dataSize),
			}

			// Заполняем данными
			for i := 0; i < tt.dataSize; i++ {
				key := strconv.Itoa(i) // Генерируем уникальный ключ
				storage.data[key] = URL{}
			}

			// Проверяем результат
			got, err := storage.URLCount()
			if err != nil {
				t.Errorf("URLCount() error = %v, wantErr false", err)
				return
			}
			if got != tt.want {
				t.Errorf("URLCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserCount(t *testing.T) {
	tests := []struct {
		name     string
		userData map[string][]string
		want     int
	}{
		{
			name:     "no users",
			userData: map[string][]string{},
			want:     0,
		},
		{
			name: "single user",
			userData: map[string][]string{
				"user1": {"url1", "url2"},
			},
			want: 1,
		},
		{
			name: "multiple users",
			userData: map[string][]string{
				"user1": {"url1"},
				"user2": {"url2"},
				"user3": {"url3"},
			},
			want: 3,
		},
		{
			name: "user without URLs",
			userData: map[string][]string{
				"user1": {},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemoryStorage{
				userURLs: tt.userData,
			}

			got, err := storage.UserCount()
			if err != nil {
				t.Errorf("UserCount() error = %v, wantErr false", err)
				return
			}
			if got != tt.want {
				t.Errorf("UserCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
