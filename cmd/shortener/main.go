//go:build !test

// Package main реализует точку входа для сервиса сокращения URL.
// Этот сервис поддерживает хранение URL в памяти, файле или базе данных, а также предоставляет HTTP API для работы с сокращенными ссылками.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener/cmd/config"
	"go.uber.org/zap"
	"golang.org/x/term"

	"github.com/mi4r/go-url-shortener/internal/compress"
	"github.com/mi4r/go-url-shortener/internal/handlers"
	"github.com/mi4r/go-url-shortener/internal/logger"
	"github.com/mi4r/go-url-shortener/internal/profiler"
	"github.com/mi4r/go-url-shortener/internal/storage"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	// buildVersion содержит версию приложения
	buildVersion string
	// buildDate содержит дату сборки приложения
	buildDate string
	// buildCommit содержит коммит сборки
	buildCommit string
)

// PrintBuildConfig выводит установленные и неустановленные флаги линковщика:
// buildVersion, buildDate, buildCommit
func PrintBuildConfig() {
	version := "N/A"
	date := "N/A"
	commit := "N/A"
	if buildVersion != "" {
		version = buildVersion
	}
	if buildDate != "" {
		date = buildDate
	}
	if buildCommit != "" {
		commit = buildCommit
	}
	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}

// isTerminal проверяет сигнал на терминальность
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// main является точкой входа в приложение. Оно выполняет следующие задачи:
// - Инициализирует логгер.
// - Загружает конфигурацию.
// - Настраивает хранилище (в памяти, файл или базу данных).
// - Регистрирует маршруты HTTP.
// - Запускает HTTP-сервер.
func main() {
	PrintBuildConfig()
	// Инициализация логгера.
	lgr, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

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

	srv := &http.Server{
		Addr:    handlers.Flags.RunAddr,
		Handler: r,
	}

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Запуск сервера.
	go func() {
		if handlers.Flags.HTTPSEnabled {
			certFile := "cert.pem"
			keyFile := "key.pem"
			logger.Sugar.Info("Starting HTTPS server", zap.String("address", handlers.Flags.RunAddr))
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Could not start server: %s\n", err)
			}
		} else {
			logger.Sugar.Info("Starting HTTP server", zap.String("address", handlers.Flags.RunAddr))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Could not start server: %s\n", err)
			}
		}
	}()

	<-shutdownSig
	logger.Sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Sugar.Fatal("Server forced to shutdown:", err)
	}

	logger.Sugar.Info("Server exited properly")
}
