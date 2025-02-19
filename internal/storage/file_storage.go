package storage

import (
	"encoding/json"
	"io"
	"os"
	"strconv"

	"github.com/mi4r/go-url-shortener/internal/logger"
)

// FileStorage представляет файловое хранилище сокращённых URL.
type FileStorage struct {
	filePath string              // Путь к файлу хранилища.
	data     map[string]URL      // Карта сокращённых URL с данными.
	userURLs map[string][]string // Карта сокращённых URL для каждого пользователя.
	nextID   int                 // Следующий уникальный идентификатор.
}

// NewFileStorage создаёт новый экземпляр файлового хранилища и загружает данные из файла.
func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		data:     make(map[string]URL),
		userURLs: make(map[string][]string),
		nextID:   1,
	}
	err := fs.loadFromFile()
	if err != nil {
		return nil, err
	}
	return fs, nil
}

// Save сохраняет URL в файловое хранилище.
func (s *FileStorage) Save(url URL) (string, error) {
	s.data[url.ShortURL] = url
	s.userURLs[url.UserID] = append(s.userURLs[url.UserID], url.ShortURL)
	s.nextID++
	return "", s.saveToFile(url)
}

// SaveBatch сохраняет пакет URL в файловое хранилище.
func (s *FileStorage) SaveBatch(urls []URL) ([]string, error) {
	ids := make([]string, 0, len(urls))

	for i := range urls {
		shortID := generateShortID()
		urls[i].ShortURL = shortID
		s.data[shortID] = urls[i]
		s.userURLs[urls[i].UserID] = append(s.userURLs[urls[i].UserID], shortID)
		s.nextID++
		ids = append(ids, shortID)
	}

	if err := s.saveBatchToFile(urls); err != nil {
		return nil, err
	}

	return ids, nil
}

// Get возвращает URL по сокращённому идентификатору.
func (s *FileStorage) Get(shortURL string) (URL, bool) {
	url, exists := s.data[shortURL]
	if !exists {
		return URL{}, false
	}
	return url, true
}

// GetURLsByUserID возвращает все URL, связанные с указанным идентификатором пользователя.
func (s *FileStorage) GetURLsByUserID(userID string) ([]URL, error) {
	shortURLs, exists := s.userURLs[userID]
	if !exists || len(shortURLs) == 0 {
		return nil, nil
	}

	var urls []URL
	for _, shortURL := range shortURLs {
		if url, found := s.data[shortURL]; found {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

// GetNextID возвращает следующий уникальный идентификатор.
func (s *FileStorage) GetNextID() (int, error) {
	return s.nextID, nil
}

// Close завершает работу файлового хранилища. Не требует особых действий.
func (s *FileStorage) Close() error {
	return nil
}

// saveToFile записывает данные URL в файл.
func (s *FileStorage) saveToFile(url URL) error {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Sugar.Error(err)
		}
	}()

	encoder := json.NewEncoder(file)
	return encoder.Encode(url)
}

// saveBatchToFile записывает пакет URL в файл.
func (s *FileStorage) saveBatchToFile(batch []URL) error {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Sugar.Error(err)
		}
	}()

	encoder := json.NewEncoder(file)
	return encoder.Encode(batch)
}

// loadFromFile загружает данные из файла в память.
func (s *FileStorage) loadFromFile() error {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Sugar.Error(err)
		}
	}()

	decoder := json.NewDecoder(file)
	for {
		var url URL
		if err := decoder.Decode(&url); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		s.data[url.ShortURL] = url
		s.userURLs[url.UserID] = append(s.userURLs[url.UserID], url.ShortURL)
		if urlID, _ := strconv.Atoi(url.CorrelationID); urlID >= s.nextID {
			s.nextID = urlID + 1
		}
	}
	logger.Sugar.Infoln(s.userURLs)
	return nil
}

// MarkURLsAsDeleted помечает указанные сокращённые URL как удалённые для указанного пользователя.
func (s *FileStorage) MarkURLsAsDeleted(userID string, ids []string) error {
	for _, id := range ids {
		if url, exists := s.data[id]; exists && url.UserID == userID {
			url.DeletedFlag = true
			s.data[id] = url
		}
	}
	return s.saveAllToFile()
}

// saveAllToFile перезаписывает файл хранилища со всеми данными.
func (s *FileStorage) saveAllToFile() error {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, url := range s.data {
		if err := encoder.Encode(url); err != nil {
			return err
		}
	}
	return nil
}

// URLCount возвращает число всех загруженных URL
func (s *FileStorage) URLCount() (int, error) {
	return len(s.data), nil
}

// UserCount возвращает количество пользователей в хранилище
func (s *FileStorage) UserCount() (int, error) {
	return len(s.userURLs), nil
}
