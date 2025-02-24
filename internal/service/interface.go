package service

import (
	"context"
	"net"

	"github.com/mi4r/go-url-shortener/internal/storage"
)

type ShortenerInterface interface {
	Shorten(ctx context.Context, originalURL, userID string) (string, error)
	GetOriginal(ctx context.Context, shortID string) (string, error)
	BatchShorten(ctx context.Context, items []storage.URL) ([]storage.URL, error)
	GetUserURLs(ctx context.Context, userID string) ([]storage.URL, error)
	DeleteUserURLs(ctx context.Context, userID string, ids []string) error
	Ping(ctx context.Context) (bool, error)
	InternalStats(ctx context.Context, ip net.IP) (urls, users int, err error)
}
