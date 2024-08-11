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

func (s *MemoryStorage) Save(url URL) error {
	s.data[url.ShortURL] = url
	s.nextID++
	return nil
}

func (s *MemoryStorage) Get(shortURL string) (URL, error) {
	url, exists := s.data[shortURL]
	if !exists {
		return URL{}, nil
	}
	return url, nil
}

func (s *MemoryStorage) GetNextID() (int, error) {
	return s.nextID, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}
