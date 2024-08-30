package storage

import "golang.org/x/exp/rand"

const (
	idLength = 8
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Storage interface {
	Save(url URL) (string, error)
	SaveBatch(urls []URL) ([]string, error)
	Get(shortURL string) (URL, bool)
	GetURLsByUserID(userID string) ([]URL, error)
	GetNextID() (int, error)
	Close() error
}

type Pinger interface {
	Ping() error
}

type URL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
	UserID        string `json:"user_id"`
	// DeletedFlag   bool   `db:"is_deleted"`
}

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
