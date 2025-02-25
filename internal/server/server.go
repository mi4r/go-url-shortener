package server

import (
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"

	"github.com/mi4r/go-url-shortener/internal/compress"
	"github.com/mi4r/go-url-shortener/internal/handlers"
	"github.com/mi4r/go-url-shortener/internal/logger"
	"github.com/mi4r/go-url-shortener/internal/profiler"
	pb "github.com/mi4r/go-url-shortener/internal/proto"
	"github.com/mi4r/go-url-shortener/internal/storage"
)

// NewRouter создаёт маршрутизатор Chi с зарегистрированными обработчиками.
func NewRouter(storage storage.Storage, trustedSubnet *net.IPNet) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware)
	r.Use(compress.CompressMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.ShortenURLHandler(storage))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.RedirectHandler(storage))
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", handlers.APIShortenURLHandler(storage))
			r.Post("/batch", handlers.BatchShortenURLHandler(storage))
		})
		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", handlers.UserURLsHandler(storage))
			r.Delete("/urls", handlers.DeleteUserURLsHandler(storage))
		})
		r.Route("/internal", func(r chi.Router) {
			r.Get("/stats", handlers.InternalStatsHandler(storage, trustedSubnet))
		})
	})

	r.Get("/ping", handlers.PingHandler(storage))
	r.Mount("/debug", profiler.Profiler())

	return r
}

// NewServer создаёт и настраивает HTTP-сервер.
func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

func NewServerGRPC(storageImpl storage.Storage, trustedSubnet *net.IPNet) *grpc.Server {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(AuthInterceptor))
	pb.RegisterShortenerServer(grpcServer, NewGRPCServer(
		storageImpl,
		handlers.Flags.BaseShortAddr,
		trustedSubnet,
	))
	return grpcServer
}

func StartGRPC(grpcServer *grpc.Server) {
	listener, err := net.Listen("tcp", handlers.Flags.GRPCAddr)
	if err != nil {
		logger.Sugar.Fatal("gRPC listen error:", err)
	}
	logger.Sugar.Info("Starting gRPC server on ", handlers.Flags.GRPCAddr)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Sugar.Fatal("gRPC serve error:", err)
	}
}
