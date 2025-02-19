//go:build !test

// Package main реализует точку входа для сервиса сокращения URL.
// Этот сервис поддерживает хранение URL в памяти, файле или базе данных, а также предоставляет HTTP API для работы с сокращенными ссылками.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/go-url-shortener/cmd/config"
	httpsconf "github.com/mi4r/go-url-shortener/cmd/https_conf"
	"go.uber.org/zap"

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

	var trustedSubnet *net.IPNet
	if handlers.Flags.TrustedSubnet != "" {
		_, subnet, err := net.ParseCIDR(handlers.Flags.TrustedSubnet)
		if err != nil {
			logger.Sugar.Fatalf("Invalid trusted subnet: %v", err)
		}
		trustedSubnet = subnet
	}

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
		r.Route("/internal", func(r chi.Router) {
			r.Get("/stats", handlers.InternalStatsHandler(storageImpl, trustedSubnet))
		})
	})
	r.Get("/ping", handlers.PingHandler(storageImpl)) // Проверка доступности хранилища.
	r.Mount("/debug", profiler.Profiler())

	srv := &http.Server{
		Addr:    handlers.Flags.RunAddr,
		Handler: r,
	}

	signalChan := httpsconf.MakeSigChan()

	// Запуск сервера.
	go func() {
		switch {
		case handlers.Flags.HTTPSEnabled:
			logger.Sugar.Info("Starting HTTPS server", zap.String("address", handlers.Flags.RunAddr))
			if err := srv.ListenAndServeTLS(httpsconf.CertFile, httpsconf.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Could not start server: %s\n", err)
			}
		default:
			logger.Sugar.Info("Starting HTTP server", zap.String("address", handlers.Flags.RunAddr))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Could not start server: %s\n", err)
			}
		}
	}()

	<-signalChan
	logger.Sugar.Info("Shutting down server...")
	storageImpl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Sugar.Fatal("Server forced to shutdown:", err)
	}

	logger.Sugar.Info("Server exited properly")
}
