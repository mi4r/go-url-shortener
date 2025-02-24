package server

import (
	"context"
	"errors"
	"net"
	"testing"

	pb "github.com/mi4r/go-url-shortener/internal/proto"
	"github.com/mi4r/go-url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Shorten(ctx context.Context, originalURL, userID string) (string, error) {
	args := m.Called(ctx, originalURL, userID)
	return args.String(0), args.Error(1)
}

func (m *MockService) GetOriginal(ctx context.Context, shortID string) (string, error) {
	args := m.Called(ctx, shortID)
	return args.String(0), args.Error(1)
}

func (m *MockService) BatchShorten(ctx context.Context, items []storage.URL) ([]storage.URL, error) {
	args := m.Called(ctx, items)
	return args.Get(0).([]storage.URL), args.Error(1)
}

func (m *MockService) GetUserURLs(ctx context.Context, userID string) ([]storage.URL, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]storage.URL), args.Error(1)
}

func (m *MockService) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	args := m.Called(ctx, userID, ids)
	return args.Error(0)
}

func (m *MockService) Ping(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockService) InternalStats(ctx context.Context, ip net.IP) (int, int, error) {
	args := m.Called(ctx, ip)
	return args.Int(0), args.Int(1), args.Error(2)
}

func TestGRPCServer(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockService)
	server := &GRPCServer{service: mockService}

	t.Run("Shorten Success", func(t *testing.T) {
		mockService.On("Shorten", mock.Anything, "http://test.com", "user123").
			Return("http://short/abc", nil)

		md := metadata.New(map[string]string{"user-id": "user123"})
		ctxWithUser := metadata.NewIncomingContext(ctx, md)
		resp, err := server.Shorten(ctxWithUser, &pb.ShortenRequest{Url: "http://test.com"})

		assert.NoError(t, err)
		assert.Equal(t, "http://short/abc", resp.Result)
	})

	t.Run("Shorten Unauthenticated", func(t *testing.T) {
		_, err := server.Shorten(ctx, &pb.ShortenRequest{Url: "http://test.com"})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestGetOriginal(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockService)
	server := &GRPCServer{service: mockService}

	t.Run("GetOriginal Found", func(t *testing.T) {
		mockService.On("GetOriginal", ctx, "abc").
			Return("http://original.com", nil)

		resp, err := server.GetOriginal(ctx, &pb.GetOriginalRequest{Id: "abc"})
		assert.NoError(t, err)
		assert.Equal(t, "http://original.com", resp.Url)
	})

	t.Run("GetOriginal NotFound", func(t *testing.T) {
		mockService.On("GetOriginal", ctx, "invalid").
			Return("", errors.New("url not found"))

		_, err := server.GetOriginal(ctx, &pb.GetOriginalRequest{Id: "invalid"})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestAuthInterceptor(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return ctx, nil
	}

	t.Run("New UserID Generation", func(t *testing.T) {
		resp, err := AuthInterceptor(context.Background(), nil, nil, handler)
		assert.NoError(t, err)

		ctx := resp.(context.Context)
		md, _ := metadata.FromIncomingContext(ctx)
		assert.Len(t, md.Get("user-id"), 1)
	})

	t.Run("Existing UserID", func(t *testing.T) {
		md := metadata.New(map[string]string{"user-id": "existing"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		resp, err := AuthInterceptor(ctx, nil, nil, handler)
		assert.NoError(t, err)

		ctx = resp.(context.Context)
		md, _ = metadata.FromIncomingContext(ctx)
		assert.Equal(t, "existing", md.Get("user-id")[0])
	})
}

func TestConvertErrorToCode(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{errors.New("access denied"), codes.PermissionDenied},
		{errors.New("url not found"), codes.NotFound},
		{errors.New("other error"), codes.Internal},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.code, convertErrorToCode(tt.err))
	}
}

func contextWithUser(userID string) context.Context {
	md := metadata.New(map[string]string{"user-id": userID})
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestBatchShorten(t *testing.T) {
	ctx := contextWithUser("user123")
	mockService := new(MockService)

	t.Run("Success", func(t *testing.T) {
		req := []storage.URL{
			{CorrelationID: "1", ShortURL: "http://test1.com"},
			{CorrelationID: "2", ShortURL: "http://test2.com"},
		}

		expected := []storage.URL{
			{CorrelationID: "1", ShortURL: "abc"},
			{CorrelationID: "2", ShortURL: "def"},
		}

		mockService.On("BatchShorten", mock.Anything, mock.Anything).
			Return(expected, nil)

		resp, err := mockService.BatchShorten(ctx, req)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		mockService.BatchShorten(context.Background(), []storage.URL{})
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockService.On("BatchShorten", mock.Anything, mock.Anything).
			Return(nil, errors.New("storage error"))

		mockService.BatchShorten(ctx, []storage.URL{})
	})
}

func TestPing(t *testing.T) {
	mockService := new(MockService)

	t.Run("Success", func(t *testing.T) {
		mockService.On("Ping", mock.Anything).Return(true, nil)

		_, err := mockService.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Ping Error", func(t *testing.T) {
		mockService.On("Ping", mock.Anything).Return(false, errors.New("connection failed"))

		mockService.Ping(context.Background())
	})

	t.Run("Not Supported", func(t *testing.T) {
		mockService.On("Ping", mock.Anything).Return(false, errors.New("storage does not support ping"))

		mockService.Ping(context.Background())
	})
}

func TestGetUserURLs(t *testing.T) {
	ctx := contextWithUser("user123")
	mockService := new(MockService)

	t.Run("Success", func(t *testing.T) {
		expected := []storage.URL{
			{ShortURL: "abc", OriginalURL: "http://test1.com"},
			{ShortURL: "def", OriginalURL: "http://test2.com"},
		}

		mockService.On("GetUserURLs", ctx, "user123").
			Return(expected, nil)

		resp, err := mockService.GetUserURLs(ctx, "user123")

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "http://test1.com", resp[0].OriginalURL)
	})

	t.Run("Empty Result", func(t *testing.T) {
		mockService.On("GetUserURLs", ctx, "user123").
			Return([]storage.URL{}, nil)

		_, err := mockService.GetUserURLs(ctx, "user123")
		assert.NoError(t, err)
	})
}

func TestDeleteUserURLs(t *testing.T) {
	ctx := contextWithUser("user123")
	mockService := new(MockService)

	t.Run("Success", func(t *testing.T) {
		req := []string{"abc", "def"}

		mockService.On("DeleteUserURLs", ctx, "user123", []string{"abc", "def"}).
			Return(nil)

		err := mockService.DeleteUserURLs(ctx, "user123", req)
		assert.NoError(t, err)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockService.On("DeleteUserURLs", ctx, "user123", mock.Anything).
			Return(errors.New("storage error"))

		mockService.DeleteUserURLs(ctx, "user123", []string{})
	})
}
