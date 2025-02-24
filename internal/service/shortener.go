package service

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"

	"github.com/mi4r/go-url-shortener/internal/storage"
)

const (
	charset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength  = 8
	batchSize = 10
)

type Shortener struct {
	Storage       storage.Storage
	BaseURL       string
	TrustedSubnet *net.IPNet
}

func NewShortener(storage storage.Storage, baseURL string, trustedSubnet *net.IPNet) *Shortener {
	return &Shortener{
		Storage:       storage,
		BaseURL:       baseURL,
		TrustedSubnet: trustedSubnet,
	}
}

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *Shortener) Shorten(ctx context.Context, originalURL, userID string) (string, error) {
	var shortID string
	for {
		shortID = generateShortID()
		if _, exists := s.Storage.Get(shortID); !exists {
			nextID, err := s.Storage.GetNextID()
			if err != nil {
				return "", fmt.Errorf("failed to generate ID: %w", err)
			}

			url := storage.URL{
				CorrelationID: strconv.Itoa(nextID),
				ShortURL:      shortID,
				OriginalURL:   originalURL,
				UserID:        userID,
			}

			existingURL, err := s.Storage.Save(url)
			if err != nil {
				return "", fmt.Errorf("storage save error: %w", err)
			}

			if existingURL != "" {
				return fmt.Sprintf("%s/%s", s.BaseURL, existingURL), nil
			}
			break
		}
	}
	return fmt.Sprintf("%s/%s", s.BaseURL, shortID), nil
}

func (s *Shortener) GetOriginal(ctx context.Context, shortID string) (string, error) {
	url, exists := s.Storage.Get(shortID)
	if !exists {
		return "", fmt.Errorf("url not found")
	}

	if url.DeletedFlag {
		return "", fmt.Errorf("url deleted")
	}

	return url.OriginalURL, nil
}

func (s *Shortener) BatchShorten(ctx context.Context, items []storage.URL) ([]storage.URL, error) {
	shortIDs, err := s.Storage.SaveBatch(items)
	if err != nil {
		return nil, fmt.Errorf("batch save failed: %w", err)
	}

	result := make([]storage.URL, len(items))
	for i, id := range shortIDs {
		result[i].ShortURL = id
	}

	return result, nil
}

func (s *Shortener) GetUserURLs(ctx context.Context, userID string) ([]storage.URL, error) {
	urls, err := s.Storage.GetURLsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("get user urls failed: %w", err)
	}
	return urls, nil
}

func (s *Shortener) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	if err := s.Storage.MarkURLsAsDeleted(userID, ids); err != nil {
		return fmt.Errorf("delete urls failed: %w", err)
	}
	return nil
}

func (s *Shortener) Ping(ctx context.Context) (bool, error) {
	if pinger, ok := s.Storage.(storage.Pinger); ok {
		return pinger.Ping() == nil, nil
	}
	return false, fmt.Errorf("storage does not support ping")
}

func (s *Shortener) InternalStats(ctx context.Context, ip net.IP) (urls, users int, err error) {
	if s.TrustedSubnet == nil || !s.TrustedSubnet.Contains(ip) {
		return 0, 0, fmt.Errorf("access denied")
	}

	urls, err = s.Storage.URLCount()
	if err != nil {
		return 0, 0, fmt.Errorf("get url count failed: %w", err)
	}

	users, err = s.Storage.UserCount()
	if err != nil {
		return 0, 0, fmt.Errorf("get user count failed: %w", err)
	}

	return urls, users, nil
}
