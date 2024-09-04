package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	Database *sql.DB
}

func MigrateDB(db *sql.DB) error {
	// Удаление дубликатов
	_, err := db.Exec(`
        DELETE FROM urls
        WHERE id NOT IN (
            SELECT MIN(id)
            FROM urls
            GROUP BY original_url
        );
    `)
	if err != nil {
		return err
	}

	// Добавление уникального индекса
	_, err = db.Exec(`
        CREATE UNIQUE INDEX IF NOT EXISTS unique_original_url_idx
        ON urls (original_url);
    `)
	return err
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	query := `
    CREATE TABLE IF NOT EXISTS urls (
        id SERIAL PRIMARY KEY,
        correlation_id VARCHAR(255) NOT NULL,
        short_url VARCHAR(255) NOT NULL UNIQUE,
        original_url TEXT NOT NULL UNIQUE,
		user_id VARCHAR(255)
    );
	ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
	`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}
	err = MigrateDB(db)
	if err != nil {
		return nil, err
	}

	return &DBStorage{Database: db}, nil
}

func (s *DBStorage) Save(url URL) (string, error) {
	_, err := s.Database.Exec("INSERT INTO urls (correlation_id, short_url, original_url, user_id) VALUES ($1, $2, $3, $4);",
		url.CorrelationID, url.ShortURL, url.OriginalURL, url.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var existingURL string
			queryErr := s.Database.QueryRow("SELECT short_url FROM urls WHERE original_url = $1;", url.OriginalURL).Scan(&existingURL)
			if queryErr != nil {
				return "", queryErr
			}
			return existingURL, nil
		}
		return "", err
	}
	return "", nil
}

func (s *DBStorage) SaveBatch(urls []URL) ([]string, error) {
	tx, err := s.Database.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO urls (correlation_id, short_url, original_url, user_id) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ids := make([]string, 0, len(urls))

	for _, url := range urls {
		var shortID string
		for {
			shortID = generateShortID()
			if err := checkUniqueShortID(tx, shortID); err == nil {
				break
			}
		}
		fmt.Println(url)
		if _, err := stmt.Exec(url.CorrelationID, shortID, url.OriginalURL, url.UserID); err != nil {
			return nil, err
		}
		ids = append(ids, shortID)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *DBStorage) Get(shortURL string) (URL, bool) {
	var url URL
	err := s.Database.QueryRow("SELECT correlation_id, short_url, original_url FROM urls WHERE short_url = $1;", shortURL).Scan(&url.CorrelationID, &url.ShortURL, &url.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return url, false
		}
		return url, false
	}
	return url, true
}

func (s *DBStorage) GetURLsByUserID(userID string) ([]URL, error) {
	rows, err := s.Database.Query("SELECT short_url, original_url FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var url URL
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (s *DBStorage) GetNextID() (int, error) {
	var nextID int
	err := s.Database.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM urls;").Scan(&nextID)
	return nextID, err
}

func (s *DBStorage) Close() error {
	return s.Database.Close()
}

func (s *DBStorage) Ping() error {
	return s.Database.Ping()
}

func checkUniqueShortID(tx *sql.Tx, shortID string) error {
	var exists bool
	err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short_url = $1)", shortID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("duplicate key value violates unique constraint")
	}
	return nil
}

func (s *DBStorage) MarkURLsAsDeleted(userID string, shortIDs []string) error {
	query := `UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2)`
	_, err := s.Database.Exec(query, userID, shortIDs)
	return err
}
