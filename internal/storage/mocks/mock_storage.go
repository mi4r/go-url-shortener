package mocks

import (
	"github.com/mi4r/go-url-shortener/internal/storage"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Get(shortID string) (storage.URL, bool) {
	args := m.Called(shortID)
	return args.Get(0).(storage.URL), args.Bool(1)
}

func (m *MockStorage) Save(url storage.URL) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetNextID() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockStorage) SaveBatch(urls []storage.URL) ([]string, error) {
	args := m.Called(urls)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) GetURLsByUserID(userID string) ([]storage.URL, error) {
	args := m.Called(userID)
	return args.Get(0).([]storage.URL), args.Error(1)
}

func (m *MockStorage) MarkURLsAsDeleted(userID string, urls []string) error {
	args := m.Called(userID, urls)
	return args.Error(0)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorage) Ping() error {
	args := m.Called()
	return args.Error(0)
}
