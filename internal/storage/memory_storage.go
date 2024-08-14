package storage

type MemoryStorage struct {
	data   map[string]URL
	nextID int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data:   make(map[string]URL),
		nextID: 1,
	}
}

func (s *MemoryStorage) Save(url URL) (string, error) {
	s.data[url.ShortURL] = url
	s.nextID++
	return "", nil
}

func (s *MemoryStorage) SaveBatch(urls []URL) ([]string, error) {
	ids := make([]string, 0, len(urls))

	for i := range urls {
		shortID := generateShortID()
		urls[i].ShortURL = shortID
		s.data[shortID] = urls[i]
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

func (s *MemoryStorage) GetNextID() (int, error) {
	return s.nextID, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}
