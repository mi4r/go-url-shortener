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
)

func main() {
	lgr, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer lgr.Sync()

	logger.Sugar = *lgr.Sugar()

	handlers.Flags = config.Init()

	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware)
	r.Use(compress.GZIPMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.ShortenURLHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.RedirectHandler)
		})
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", handlers.APIShortenURLHandler)
	})

	logger.Sugar.Info("Starting server", zap.String("address", handlers.Flags.RunAddr))
	log.Fatal(http.ListenAndServe(handlers.Flags.RunAddr, r))
}
