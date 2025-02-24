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

	"github.com/mi4r/go-url-shortener/cmd/config"
	httpsconf "github.com/mi4r/go-url-shortener/cmd/https_conf"
	pb "github.com/mi4r/go-url-shortener/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/mi4r/go-url-shortener/internal/handlers"
	"github.com/mi4r/go-url-shortener/internal/logger"
	"github.com/mi4r/go-url-shortener/internal/server"
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
	r := server.NewRouter(storageImpl, trustedSubnet)
	srv := server.NewServer(handlers.Flags.RunAddr, r)

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

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(server.AuthInterceptor))
	pb.RegisterShortenerServer(grpcServer, server.NewGRPCServer(
		storageImpl,
		handlers.Flags.BaseShortAddr,
		trustedSubnet,
	))

	go func() {
		listener, err := net.Listen("tcp", handlers.Flags.GRPCAddr)
		if err != nil {
			logger.Sugar.Fatal("gRPC listen error:", err)
		}
		logger.Sugar.Info("Starting gRPC server on ", handlers.Flags.GRPCAddr)
		if err := grpcServer.Serve(listener); err != nil {
			logger.Sugar.Fatal("gRPC serve error:", err)
		}
	}()

	<-signalChan
	logger.Sugar.Info("Shutting down server...")
	storageImpl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	grpcServer.GracefulStop()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Sugar.Fatal("Server forced to shutdown:", err)
	}

	logger.Sugar.Info("Server exited properly")
}
