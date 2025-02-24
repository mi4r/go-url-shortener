// internal/server/grpc.go
package server

import (
	"context"
	"fmt"
	"net"
	"strings"

	pb "github.com/mi4r/go-url-shortener/internal/proto"
	"github.com/mi4r/go-url-shortener/internal/service"
	"github.com/mi4r/go-url-shortener/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	pb.UnimplementedShortenerServer
	service *service.Shortener
}

func NewGRPCServer(storage storage.Storage, baseURL string, trustedSubnet *net.IPNet) *GRPCServer {
	return &GRPCServer{
		service: service.NewShortener(storage, baseURL, trustedSubnet),
	}
}

func (s *GRPCServer) Shorten(ctx context.Context, req *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}

	shortURL, err := s.service.Shorten(ctx, req.GetUrl(), userID)
	if err != nil {
		return nil, status.Error(convertErrorToCode(err), err.Error())
	}

	return &pb.ShortenResponse{
		Result: shortURL,
	}, nil
}

func (s *GRPCServer) GetOriginal(ctx context.Context, req *pb.GetOriginalRequest) (*pb.GetOriginalResponse, error) {
	originalURL, err := s.service.GetOriginal(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(convertErrorToCode(err), err.Error())
	}

	return &pb.GetOriginalResponse{
		Url: originalURL,
	}, nil
}

func (s *GRPCServer) BatchShorten(ctx context.Context, req *pb.BatchShortenRequest) (*pb.BatchShortenResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}

	items := make([]storage.URL, len(req.GetItems()))
	for i, item := range req.GetItems() {
		items[i] = storage.URL{
			CorrelationID: item.GetCorrelationId(),
			OriginalURL:   item.GetOriginalUrl(),
			UserID:        userID,
		}
	}

	result, err := s.service.BatchShorten(ctx, items)
	if err != nil {
		return nil, status.Error(convertErrorToCode(err), err.Error())
	}

	responseItems := make([]*pb.BatchShortenResponseItem, len(result))
	for i, item := range result {
		responseItems[i] = &pb.BatchShortenResponseItem{
			CorrelationId: item.CorrelationID,
			ShortUrl:      fmt.Sprintf("%s/%s", s.service.BaseURL, item.ShortURL),
		}
	}

	return &pb.BatchShortenResponse{Items: responseItems}, nil
}

func (s *GRPCServer) GetUserURLs(ctx context.Context, _ *pb.Empty) (*pb.GetUserURLsResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}

	urls, err := s.service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(convertErrorToCode(err), err.Error())
	}

	responseItems := make([]*pb.URLResponseItem, len(urls))
	for i, url := range urls {
		responseItems[i] = &pb.URLResponseItem{
			ShortUrl:    fmt.Sprintf("%s/%s", s.service.BaseURL, url.ShortURL),
			OriginalUrl: url.OriginalURL,
		}
	}

	return &pb.GetUserURLsResponse{Items: responseItems}, nil
}

func (s *GRPCServer) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.Empty, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}

	if err := s.service.DeleteUserURLs(ctx, userID, req.GetIds()); err != nil {
		return nil, status.Error(convertErrorToCode(err), err.Error())
	}

	return nil, nil
}

func (s *GRPCServer) Ping(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	ok, err := s.service.Ping(ctx)
	if !ok || err != nil {
		return nil, status.Error(convertErrorToCode(err), err.Error())
	}

	return nil, nil
}

func (s *GRPCServer) InternalStats(ctx context.Context, req *pb.InternalStatsRequest) (*pb.InternalStatsResponse, error) {
	p, _ := peer.FromContext(ctx)
	ip := net.ParseIP(strings.Split(p.Addr.String(), ":")[0])

	urls, users, err := s.service.InternalStats(ctx, ip)
	if err != nil {
		return nil, status.Error(convertErrorToCode(err), err.Error())
	}

	return &pb.InternalStatsResponse{
		UrlsCnt:  int32(urls),
		UsersCnt: int32(users),
	}, nil
}

func getUserIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("user-id")
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func convertErrorToCode(err error) codes.Code {
	switch {
	case strings.Contains(err.Error(), "url not found"):
		return codes.NotFound
	case strings.Contains(err.Error(), "url deleted"):
		return codes.NotFound
	case strings.Contains(err.Error(), "access denied"):
		return codes.PermissionDenied
	case strings.Contains(err.Error(), "missing user ID"):
		return codes.Unauthenticated
	default:
		return codes.Internal
	}
}
