package storage

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	Database *sql.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	query := `
    CREATE TABLE IF NOT EXISTS urls (
        id SERIAL PRIMARY KEY,
        uuid VARCHAR(255) NOT NULL,
        short_url VARCHAR(255) NOT NULL UNIQUE,
        original_url TEXT NOT NULL
    );`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &DBStorage{Database: db}, nil
}

func (s *DBStorage) Save(url URL) error {
	_, err := s.Database.Exec("INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)", url.UUID, url.ShortURL, url.OriginalURL)
	return err
}

func (s *DBStorage) Get(shortURL string) (URL, error) {
	var url URL
	err := s.Database.QueryRow("SELECT uuid, short_url, original_url FROM urls WHERE short_url = $1", shortURL).Scan(&url.UUID, &url.ShortURL, &url.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return url, nil
		}
		return url, err
	}
	return url, nil
}

func (s *DBStorage) GetNextID() (int, error) {
	var nextID int
	err := s.Database.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM urls").Scan(&nextID)
	return nextID, err
}

func (s *DBStorage) Close() error {
	return s.Database.Close()
}

func (s *DBStorage) Ping() error {
	return s.Database.Ping()
}