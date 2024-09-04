package storage

import (
	"encoding/json"
	"io"
	"os"
	"strconv"

	"github.com/mi4r/go-url-shortener/internal/logger"
)

type FileStorage struct {
	filePath string
	data     map[string]URL
	userURLs map[string][]string
	nextID   int
}

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

func (s *FileStorage) Save(url URL) (string, error) {
	s.data[url.ShortURL] = url
	s.userURLs[url.UserID] = append(s.userURLs[url.UserID], url.ShortURL)
	s.nextID++
	return "", s.saveToFile(url)
}

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

func (s *FileStorage) Get(shortURL string) (URL, bool) {
	url, exists := s.data[shortURL]
	if !exists {
		return URL{}, false
	}
	return url, true
}

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

func (s *FileStorage) GetNextID() (int, error) {
	return s.nextID, nil
}

func (s *FileStorage) Close() error {
	return nil
}

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

func (s *FileStorage) MarkURLsAsDeleted(userID string, ids []string) error {
	for _, id := range ids {
		if url, exists := s.data[id]; exists && url.UserID == userID {
			url.DeletedFlag = true
			s.data[id] = url
		}
	}
	return s.saveAllToFile()
}

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
