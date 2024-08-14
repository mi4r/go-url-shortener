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
	GetNextID() (int, error)
	Close() error
}

type Pinger interface {
	Ping() error
}

type URL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
