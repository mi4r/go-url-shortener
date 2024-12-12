package storage

import (
	"os"
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
