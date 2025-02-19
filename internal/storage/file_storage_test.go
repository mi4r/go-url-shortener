package storage

import (
	"os"
	"strconv"
	"testing"

	"github.com/mi4r/go-url-shortener/internal/logger"
	"go.uber.org/zap"
)

func TestFileStorage(t *testing.T) {
	// Инициализация временного файла для тестов
	tempFile, err := os.CreateTemp("", "storage_test_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Инициализация логгера
	logger.Sugar = *zap.NewNop().Sugar()

	// Создание нового хранилища
	fs, err := NewFileStorage(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	// Тесты Save
	t.Run("Save", func(t *testing.T) {
		url := URL{
			CorrelationID: "1",
			ShortURL:      "short1",
			OriginalURL:   "https://example.com",
			UserID:        "user1",
		}
		_, err := fs.Save(url)
		if err != nil {
			t.Errorf("failed to save URL: %v", err)
		}
	})

	// Тесты SaveBatch
	t.Run("SaveBatch", func(t *testing.T) {
		urls := []URL{
			{CorrelationID: "2", OriginalURL: "https://example2.com", UserID: "user1"},
			{CorrelationID: "3", OriginalURL: "https://example3.com", UserID: "user1"},
		}
		ids, err := fs.SaveBatch(urls)
		if err != nil {
			t.Errorf("failed to save batch: %v", err)
		}
		if len(ids) != len(urls) {
			t.Errorf("expected %d IDs, got %d", len(urls), len(ids))
		}
	})

	// Тесты Get
	t.Run("Get", func(t *testing.T) {
		url, exists := fs.Get("short1")
		if !exists {
			t.Errorf("failed to get URL")
		}
		if url.OriginalURL != "https://example.com" {
			t.Errorf("expected https://example.com, got %s", url.OriginalURL)
		}
	})

	// Тесты GetURLsByUserID
	t.Run("GetURLsByUserID", func(t *testing.T) {
		urls, err := fs.GetURLsByUserID("user1")
		if err != nil {
			t.Errorf("failed to get URLs by user ID: %v", err)
		}
		if len(urls) < 1 {
			t.Errorf("expected at least 1 URL, got %d", len(urls))
		}
	})

	// Тесты GetNextID
	t.Run("GetNextID", func(t *testing.T) {
		nextID, err := fs.GetNextID()
		if err != nil {
			t.Errorf("failed to get next ID: %v", err)
		}
		if nextID <= 0 {
			t.Errorf("invalid next ID: %d", nextID)
		}
	})

	// Тесты MarkURLsAsDeleted
	t.Run("MarkURLsAsDeleted", func(t *testing.T) {
		err := fs.MarkURLsAsDeleted("user1", []string{"short1"})
		if err != nil {
			t.Errorf("failed to mark URLs as deleted: %v", err)
		}
		url, exists := fs.Get("short1")
		if !exists || !url.DeletedFlag {
			t.Errorf("expected URL to be marked as deleted")
		}
	})

	// Тесты Close
	t.Run("Close", func(t *testing.T) {
		err := fs.Close()
		if err != nil {
			t.Errorf("failed to close storage: %v", err)
		}
	})
}

func TestFileURLCount(t *testing.T) {
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
			storage := &FileStorage{
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

func TestFileUserCount(t *testing.T) {
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
			storage := &FileStorage{
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
