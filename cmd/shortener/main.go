//go:build !test

// Package main реализует точку входа для сервиса сокращения URL.
// Этот сервис поддерживает хранение URL в памяти, файле или базе данных, а также предоставляет HTTP API для работы с сокращенными ссылками.
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
	"github.com/mi4r/go-url-shortener/internal/profiler"
	"github.com/mi4r/go-url-shortener/internal/storage"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// main является точкой входа в приложение. Оно выполняет следующие задачи:
// - Инициализирует логгер.
// - Загружает конфигурацию.
// - Настраивает хранилище (в памяти, файл или базу данных).
// - Регистрирует маршруты HTTP.
// - Запускает HTTP-сервер.
func main() {
	// Инициализация логгера.
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

	// Загрузка конфигурации.
	handlers.Flags = config.Init()

	var storageImpl storage.Storage

	// Настройка хранилища с приоритетом: база данных > файл > память.
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

	// Инициализация маршрутизатора.
	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware)    // Логирование запросов.
	r.Use(compress.CompressMiddleware) // Сжатие ответов.

	// Основные маршруты API.
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.ShortenURLHandler(storageImpl)) // Сокращение URL.
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.RedirectHandler(storageImpl)) // Редирект по сокращенному URL.
		})
	})
	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", handlers.APIShortenURLHandler(storageImpl))        // Сокращение URL в формате JSON.
			r.Post("/batch", handlers.BatchShortenURLHandler(storageImpl)) // API для пакетного сокращения URL.
		})
		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", handlers.UserURLsHandler(storageImpl))          // Получение всех URL пользователя.
			r.Delete("/urls", handlers.DeleteUserURLsHandler(storageImpl)) // Удаление URL пользователя.
		})
	})
	r.Get("/ping", handlers.PingHandler(storageImpl)) // Проверка доступности хранилища.
	r.Mount("/debug", profiler.Profiler())

	// Запуск HTTP-сервера.
	logger.Sugar.Info("Starting server", zap.String("address", handlers.Flags.RunAddr))
	log.Fatal(http.ListenAndServe(handlers.Flags.RunAddr, r))
}
