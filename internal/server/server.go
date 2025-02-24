package server

import (
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mi4r/go-url-shortener/internal/compress"
	"github.com/mi4r/go-url-shortener/internal/handlers"
	"github.com/mi4r/go-url-shortener/internal/logger"
	"github.com/mi4r/go-url-shortener/internal/profiler"
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
