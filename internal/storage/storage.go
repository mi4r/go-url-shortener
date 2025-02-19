package storage

import "golang.org/x/exp/rand"

// Константы для генерации короткого идентификатора.
const (
	idLength = 8                                                                // Длина короткого идентификатора.
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // Набор символов для идентификатора.
)

// Storage определяет интерфейс для работы с хранилищем URL.
type Storage interface {
	// Save сохраняет URL в хранилище.
	Save(url URL) (string, error)
	// SaveBatch сохраняет пакет URL в хранилище.
	SaveBatch(urls []URL) ([]string, error)
	// Get возвращает URL по короткому идентификатору.
	Get(shortURL string) (URL, bool)
	// GetURLsByUserID возвращает все URL, связанные с указанным идентификатором пользователя.
	GetURLsByUserID(userID string) ([]URL, error)
	// GetNextID возвращает следующий уникальный идентификатор для новой записи.
	GetNextID() (int, error)
	// Close закрывает хранилище.
	Close() error
	// MarkURLsAsDeleted помечает список URL как удаленные для указанного пользователя.
	MarkURLsAsDeleted(userID string, shortIDs []string) error
	// URLCount возвращает число всех загруженных URL
	URLCount() (int, error)
	// UserCount возвращает количество пользователей в хранилище
	UserCount() (int, error)
}

// Pinger определяет интерфейс для проверки доступности соединения.
type Pinger interface {
	// Ping проверяет доступность соединения с хранилищем.
	Ping() error
}

// URL представляет структуру данных для хранения информации об URL.
type URL struct {
	CorrelationID string `json:"correlation_id"` // Корреляционный идентификатор.
	ShortURL      string `json:"short_url"`      // Короткий URL.
	OriginalURL   string `json:"original_url"`   // Оригинальный URL.
	UserID        string `json:"user_id"`        // Идентификатор пользователя.
	DeletedFlag   bool   `json:"is_deleted"`     // Флаг удаления URL.
}

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
