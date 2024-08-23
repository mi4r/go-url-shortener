package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener/cmd/config"
	"go.uber.org/zap"

	"github.com/mi4r/go-url-shortener/internal/compress"
	"github.com/mi4r/go-url-shortener/internal/handlers"
	"github.com/mi4r/go-url-shortener/internal/logger"
	"github.com/mi4r/go-url-shortener/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	lgr, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func() {
		if err := lgr.Sync(); err != nil {
			logger.Sugar.Error(err)
		}
	}()

	logger.Sugar = *lgr.Sugar()

	handlers.Flags = config.Init()

	var storageImpl storage.Storage

	if handlers.Flags.DataBaseDSN != "" {
		storageImpl, err = storage.NewDBStorage(handlers.Flags.DataBaseDSN)
		if err != nil {
			logger.Sugar.Warn("Falling back to file storage due to DB error: ", err)
		}
	}
	if storageImpl == nil && handlers.Flags.URLStorageFilePath != "" {
		storageImpl, err = storage.NewFileStorage(handlers.Flags.URLStorageFilePath)
		if err != nil {
			logger.Sugar.Warn("Falling back to memory storage due to file error: ", err)
		}
	}
	if storageImpl == nil {
		storageImpl = storage.NewMemoryStorage()
		logger.Sugar.Info("Using in-memory storage")
	}
	defer storageImpl.Close()

	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware)
	r.Use(compress.CompressMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.ShortenURLHandler(storageImpl))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.RedirectHandler(storageImpl))
		})
	})
	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", handlers.APIShortenURLHandler(storageImpl))
			r.Post("/batch", handlers.BatchShortenURLHandler(storageImpl))
		})
		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", handlers.UserURLsHandler(storageImpl))
		})
	})
	r.Get("/ping", handlers.PingHandler(storageImpl))

	logger.Sugar.Info("Starting server", zap.String("address", handlers.Flags.RunAddr))
	log.Fatal(http.ListenAndServe(handlers.Flags.RunAddr, r))
}
