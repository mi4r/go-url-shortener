package storage

import "github.com/mi4r/go-url-shortener/internal/logger"

type MemoryStorage struct {
	data     map[string]URL
	userURLs map[string][]string
	nextID   int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data:     make(map[string]URL),
		userURLs: make(map[string][]string),
		nextID:   1,
	}
}

func (s *MemoryStorage) Save(url URL) (string, error) {
	s.data[url.ShortURL] = url
	s.userURLs[url.UserID] = append(s.userURLs[url.UserID], url.ShortURL)
	s.nextID++
	return "", nil
}

func (s *MemoryStorage) SaveBatch(urls []URL) ([]string, error) {
	ids := make([]string, 0, len(urls))

	for i := range urls {
		shortID := generateShortID()
		urls[i].ShortURL = shortID
		s.data[shortID] = urls[i]
		s.userURLs[urls[i].UserID] = append(s.userURLs[urls[i].UserID], shortID)
		s.nextID++
		ids = append(ids, shortID)
	}

	return ids, nil
}

func (s *MemoryStorage) Get(shortURL string) (URL, bool) {
	url, exists := s.data[shortURL]
	if !exists {
		return URL{}, false
	}
	return url, true
}

func (s *MemoryStorage) GetURLsByUserID(userID string) ([]URL, error) {
	var urls []URL
	shortIDs, exists := s.userURLs[userID]
	if !exists {
		return nil, nil
	}

	for _, shortID := range shortIDs {
		url, ok := s.data[shortID]
		if ok {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

func (s *MemoryStorage) GetNextID() (int, error) {
	return s.nextID, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) MarkURLsAsDeleted(userID string, ids []string) error {
	for _, id := range ids {
		url, exists := s.data[id]
		if exists && url.UserID == userID {
			url.DeletedFlag = true
			s.data[id] = url
		}
	}
	logger.Sugar.Info(s.data)
	return nil
}
