package storage

type Storage interface {
	Save(url URL) error
	Get(shortURL string) (URL, error)
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
