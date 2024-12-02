package storage_test

import (
	"fmt"
	"testing"

	"github.com/mi4r/go-url-shortener/cmd/config"
	"github.com/mi4r/go-url-shortener/internal/handlers"
	"github.com/mi4r/go-url-shortener/internal/storage"
)

func BenchmarkSaveURL(b *testing.B) {
	handlers.Flags = &config.Flags{
		RunAddr:            "localhost:8080",
		BaseShortAddr:      "http://localhost:8080",
		URLStorageFilePath: "test_storage.json",
		DataBaseDSN:        "host=localhost user=url_storage password=1234 dbname=url_storage sslmode=disable",
	}
	store := storage.NewMemoryStorage()
	defer store.Close()

	// Генерация URL
	url := storage.URL{
		ShortURL:    "shortURL",
		OriginalURL: "http://example.com",
		UserID:      "test_user",
	}

	b.ResetTimer() // Сбрасываем таймер перед началом тестов

	for i := 0; i < b.N; i++ {
		url.ShortURL = fmt.Sprintf("shortURL_%d", i)
		if _, err := store.Save(url); err != nil {
			b.Errorf("failed to save URL: %v", err)
		}
	}
}
