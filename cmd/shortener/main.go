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

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	lgr, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer lgr.Sync()

	logger.Sugar = *lgr.Sugar()

	handlers.Flags = config.Init()

	err = handlers.LoadFromFile(handlers.Flags.URLStorageFilePath)
	if err != nil {
		log.Printf("Failed to load data from file: %v", err)
	}

	handlers.Database, err = sql.Open("pgx", handlers.Flags.DataBaseDSN)
	if err != nil {
		logger.Sugar.Error("Cannot open database", err)
	}
	defer handlers.Database.Close()

	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware)
	r.Use(compress.CompressMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.ShortenURLHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.RedirectHandler)
		})
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", handlers.APIShortenURLHandler)
	})
	r.Get("/ping", handlers.PingHandler)

	logger.Sugar.Info("Starting server", zap.String("address", handlers.Flags.RunAddr))
	log.Fatal(http.ListenAndServe(handlers.Flags.RunAddr, r))
}
