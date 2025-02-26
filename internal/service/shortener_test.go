package service

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/mi4r/go-url-shortener/internal/storage"
	"github.com/mi4r/go-url-shortener/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortener_Shorten(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("Get", mock.AnythingOfType("string")).Return(storage.URL{}, false)
		mockStorage.On("GetNextID").Return(1, nil)
		mockStorage.On("Save", mock.AnythingOfType("storage.URL")).Return("", nil)

		s := NewShortener(mockStorage, "http://short", nil)
		result, err := s.Shorten(context.Background(), "http://original", "user1")

		assert.NoError(t, err)
		assert.Regexp(t, `^http://short/[a-zA-Z0-9]{8}$`, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("existing url", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("Get", mock.AnythingOfType("string")).Return(storage.URL{}, false)
		mockStorage.On("GetNextID").Return(1, nil)
		mockStorage.On("Save", mock.Anything).Return("existingID", nil)

		s := NewShortener(mockStorage, "http://short", nil)
		result, err := s.Shorten(context.Background(), "http://original", "user1")

		assert.NoError(t, err)
		assert.Equal(t, "http://short/existingID", result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("Get", mock.AnythingOfType("string")).Return(storage.URL{}, false)
		mockStorage.On("GetNextID").Return(0, errors.New("id error"))

		s := NewShortener(mockStorage, "http://short", nil)
		_, err := s.Shorten(context.Background(), "http://original", "user1")

		assert.ErrorContains(t, err, "failed to generate ID")
		mockStorage.AssertExpectations(t)
	})
}

func TestShortener_GetOriginal(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("Get", "valid").Return(storage.URL{OriginalURL: "http://original"}, true)

		s := NewShortener(mockStorage, "", nil)
		result, err := s.GetOriginal(context.Background(), "valid")

		assert.NoError(t, err)
		assert.Equal(t, "http://original", result)
	})

	t.Run("not found", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("Get", "invalid").Return(storage.URL{}, false)

		s := NewShortener(mockStorage, "", nil)
		_, err := s.GetOriginal(context.Background(), "invalid")

		assert.ErrorContains(t, err, "url not found")
	})

	t.Run("deleted", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("Get", "deleted").Return(storage.URL{DeletedFlag: true}, true)

		s := NewShortener(mockStorage, "", nil)
		_, err := s.GetOriginal(context.Background(), "deleted")

		assert.ErrorContains(t, err, "url deleted")
	})
}

func TestShortener_BatchShorten(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("SaveBatch", mock.Anything).Return([]string{"id1", "id2"}, nil)

		s := NewShortener(mockStorage, "http://short", nil)
		urls := []storage.URL{
			{OriginalURL: "http://1"},
			{OriginalURL: "http://2"},
		}

		result, err := s.BatchShorten(context.Background(), urls)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "id1", result[0].ShortURL)
	})
}

func TestShortener_GetUserURLs(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("GetURLsByUserID", "user1").Return([]storage.URL{
			{ShortURL: "id1", OriginalURL: "http://1"},
		}, nil)

		s := NewShortener(mockStorage, "http://short", nil)
		result, err := s.GetUserURLs(context.Background(), "user1")

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "id1", result[0].ShortURL)
	})

	t.Run("empty", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("GetURLsByUserID", "user2").Return([]storage.URL{}, nil)

		s := NewShortener(mockStorage, "http://short", nil)
		result, err := s.GetUserURLs(context.Background(), "user2")

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestShortener_DeleteUserURLs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("MarkURLsAsDeleted", "user1", []string{"id1"}).Return(nil)

		s := NewShortener(mockStorage, "", nil)
		err := s.DeleteUserURLs(context.Background(), "user1", []string{"id1"})

		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("MarkURLsAsDeleted", "user1", mock.Anything).Return(errors.New("delete error"))

		s := NewShortener(mockStorage, "", nil)
		err := s.DeleteUserURLs(context.Background(), "user1", []string{"id1"})

		assert.ErrorContains(t, err, "delete urls failed")
	})
}

func TestShortener_Ping(t *testing.T) {
	t.Run("pinger ok", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("Ping").Return(nil)

		s := NewShortener(mockStorage, "", nil)
		ok, err := s.Ping(context.Background())

		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestShortener_InternalStats(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.0.0/24")

	t.Run("access denied", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		s := NewShortener(mockStorage, "", subnet)

		_, _, err := s.InternalStats(context.Background(), net.ParseIP("10.0.0.1"))
		assert.ErrorContains(t, err, "access denied")
	})

	t.Run("success", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("URLCount").Return(10, nil)
		mockStorage.On("UserCount").Return(5, nil)

		s := NewShortener(mockStorage, "", subnet)
		urls, users, err := s.InternalStats(context.Background(), net.ParseIP("192.168.0.1"))

		assert.NoError(t, err)
		assert.Equal(t, 10, urls)
		assert.Equal(t, 5, users)
	})
}
