// Package mocks provides a mock implementation of the storage.Storage interface
// for use in unit tests. This mock is built using the github.com/stretchr/testify/mock
// library, allowing developers to simulate storage behavior without relying on
// actual storage implementations.
package mocks

import (
	"github.com/mi4r/go-url-shortener/internal/storage"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the storage.Storage interface.
// It extends the mock.Mock type from the testify/mock package, enabling
// developers to set expectations and simulate storage operations in tests.
type MockStorage struct {
	mock.Mock
}

// Get retrieves a URL by its short ID. This method is a mock implementation
// and can be configured to return specific values or errors during tests.
//
// Parameters:
//   - shortID: The short ID of the URL to retrieve.
//
// Returns:
//   - storage.URL: The URL object corresponding to the short ID.
//   - bool: True if the URL exists, false otherwise.
func (m *MockStorage) Get(shortID string) (storage.URL, bool) {
	args := m.Called(shortID)
	return args.Get(0).(storage.URL), args.Bool(1)
}

// Save stores a URL in the mock storage. This method is a mock implementation
// and can be configured to return specific short IDs or errors during tests.
//
// Parameters:
//   - url: The URL object to be saved.
//
// Returns:
//   - string: The short ID generated for the URL.
//   - error: An error if the save operation fails.
func (m *MockStorage) Save(url storage.URL) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

// GetNextID retrieves the next available ID for storing a URL. This method
// is a mock implementation and can be configured to return specific IDs or
// errors during tests.
//
// Returns:
//   - int: The next available ID.
//   - error: An error if the operation fails.
func (m *MockStorage) GetNextID() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// SaveBatch stores a batch of URLs in the mock storage. This method is a mock
// implementation and can be configured to return specific short IDs or errors
// during tests.
//
// Parameters:
//   - urls: A slice of URL objects to be saved.
//
// Returns:
//   - []string: A slice of short IDs generated for the URLs.
//   - error: An error if the save operation fails.
func (m *MockStorage) SaveBatch(urls []storage.URL) ([]string, error) {
	args := m.Called(urls)
	return args.Get(0).([]string), args.Error(1)
}

// GetURLsByUserID retrieves all URLs associated with a specific user ID.
// This method is a mock implementation and can be configured to return specific
// URL lists or errors during tests.
//
// Parameters:
//   - userID: The ID of the user whose URLs are to be retrieved.
//
// Returns:
//   - []storage.URL: A slice of URL objects associated with the user.
//   - error: An error if the retrieval operation fails.
func (m *MockStorage) GetURLsByUserID(userID string) ([]storage.URL, error) {
	args := m.Called(userID)
	return args.Get(0).([]storage.URL), args.Error(1)
}

// MarkURLsAsDeleted marks a list of URLs as deleted for a specific user.
// This method is a mock implementation and can be configured to return specific
// errors during tests.
//
// Parameters:
//   - userID: The ID of the user whose URLs are to be marked as deleted.
//   - urls: A slice of short IDs of the URLs to be marked as deleted.
//
// Returns:
//   - error: An error if the operation fails.
func (m *MockStorage) MarkURLsAsDeleted(userID string, urls []string) error {
	args := m.Called(userID, urls)
	return args.Error(0)
}

// Close closes the mock storage. This method is a mock implementation and
// can be configured to return specific errors during tests.
//
// Returns:
//   - error: An error if the close operation fails.
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Ping checks the connection to the mock storage. This method is a mock
// implementation and can be configured to return specific errors during tests.
//
// Returns:
//   - error: An error if the ping operation fails.
func (m *MockStorage) Ping() error {
	args := m.Called()
	return args.Error(0)
}

// URLCount returns count of shorten URLs
//
// Returns:
//
//	-int: number of URLs
//	- error: An error if the operation fails.
func (m *MockStorage) URLCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// UserCount returns count of users
//
// Returns:
//
//	-int: number of users
//	- error: An error if the operation fails.
func (m *MockStorage) UserCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}
