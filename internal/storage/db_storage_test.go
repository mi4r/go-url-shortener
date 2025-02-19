package storage

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mi4r/go-url-shortener/internal/logger"
	"go.uber.org/zap"
)

const testDSN = "postgresql://mi4r:1234@localhost/test_db?sslmode=disable"

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", testDSN)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS urls;`)
	if err != nil {
		t.Fatalf("failed to clean test database: %v", err)
	}

	_, err = db.Exec(`
    CREATE TABLE urls (
        id SERIAL PRIMARY KEY,
        correlation_id VARCHAR(255) NOT NULL,
        short_url VARCHAR(255) NOT NULL UNIQUE,
        original_url TEXT NOT NULL UNIQUE,
        user_id VARCHAR(255),
        is_deleted BOOLEAN DEFAULT FALSE
    );`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	return db
}

func TestDBStorage(t *testing.T) {
	logger.Sugar = *zap.NewNop().Sugar()
	db := setupTestDB(t)
	defer db.Close()

	storage, err := NewDBStorage(testDSN)
	if err := storage.Ping(); err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
	}
	if err != nil {
		t.Fatalf("failed to initialize DBStorage: %v", err)
	}
	defer storage.Close()

	t.Run("Save", func(t *testing.T) {
		url := URL{
			CorrelationID: "1",
			ShortURL:      "short1",
			OriginalURL:   "https://example.com",
			UserID:        "user1",
		}

		_, err := storage.Save(url)
		if err != nil {
			t.Errorf("Save failed: %v", err)
		}

		savedURL, exists := storage.Get("short1")
		if !exists {
			t.Errorf("expected URL to exist")
		}
		if savedURL.OriginalURL != url.OriginalURL {
			t.Errorf("expected %s, got %s", url.OriginalURL, savedURL.OriginalURL)
		}
	})

	t.Run("SaveBatch", func(t *testing.T) {
		urls := []URL{
			{CorrelationID: "2", OriginalURL: "https://example2.com", UserID: "user1"},
			{CorrelationID: "3", OriginalURL: "https://example3.com", UserID: "user1"},
		}

		ids, err := storage.SaveBatch(urls)
		if err != nil {
			t.Errorf("SaveBatch failed: %v", err)
		}

		if len(ids) != len(urls) {
			t.Errorf("expected %d IDs, got %d", len(urls), len(ids))
		}
	})

	t.Run("Get", func(t *testing.T) {
		url, exists := storage.Get("short1")
		if !exists {
			t.Errorf("Get failed: expected URL to exist")
		}
		if url.OriginalURL != "https://example.com" {
			t.Errorf("expected %s, got %s", "https://example.com", url.OriginalURL)
		}
	})

	t.Run("GetURLsByUserID", func(t *testing.T) {
		urls, err := storage.GetURLsByUserID("user1")
		if err != nil {
			t.Errorf("GetURLsByUserID failed: %v", err)
		}
		if len(urls) < 1 {
			t.Errorf("expected at least 1 URL, got %d", len(urls))
		}
	})

	t.Run("GetNextID", func(t *testing.T) {
		nextID, err := storage.GetNextID()
		if err != nil {
			t.Errorf("GetNextID failed: %v", err)
		}
		if nextID <= 0 {
			t.Errorf("invalid next ID: %d", nextID)
		}
	})

	t.Run("MarkURLsAsDeleted", func(t *testing.T) {
		err := storage.MarkURLsAsDeleted("user1", []string{"short1"})
		if err != nil {
			t.Errorf("MarkURLsAsDeleted failed: %v", err)
		}

		// err = storage.MarkURLsAsDeleted("user2", []string{"short1", "bad"})
		// if err == nil {
		// 	t.Errorf("MarkURLsAsDeleted should be failed")
		// }

		url, exists := storage.Get("short1")
		if !exists || !url.DeletedFlag {
			t.Errorf("expected URL to be marked as deleted")
		}
	})

	t.Run("Ping", func(t *testing.T) {
		err := storage.Ping()
		if err != nil {
			t.Errorf("Ping failed: %v", err)
		}
	})
}
