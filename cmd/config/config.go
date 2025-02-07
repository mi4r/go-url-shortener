// Package config предоставляет функционал для конфигурации приложения.
// Оно позволяет инициализировать параметры через флаги командной строки,
// переменные окружения или использовать значения по умолчанию.
package config

import (
	"flag"
	"fmt"
	"os"
)

// Flags представляет конфигурационные параметры приложения.
type Flags struct {
	RunAddr            string // Адрес и порт для запуска сервера.
	BaseShortAddr      string // Базовый URL для сокращенных ссылок.
	URLStorageFilePath string // Путь к файлу для хранения URL (если используется файловое хранилище).
	DataBaseDSN        string // DSN (Data Source Name) для подключения к базе данных.
	HTTPSEnabled       bool   // Возможность подключения к HTTPS-серверу
}

// String возвращает строковое представление текущих параметров конфигурации.
func (f *Flags) String() string {
	return fmt.Sprintf("RunAddr: %s, BaseShortAddr: %s, URLStorageFileName: %s, DataBaseDSN: %s", f.RunAddr, f.BaseShortAddr, f.URLStorageFilePath, f.DataBaseDSN)
}

// Init инициализирует параметры конфигурации из флагов командной строки, переменных окружения и значений по умолчанию.
// Порядок приоритета:
// 1. Переменные окружения.
// 2. Флаги командной строки.
// 3. Значения по умолчанию.
func Init() *Flags {
	// Определение флагов командной строки с их значениями по умолчанию и описанием.
	addr := flag.String("a", "localhost:8080", "Address and port to run server")
	base := flag.String("b", "http://localhost:8080", "Base shorten url")
	storagePath := flag.String("f", "", "URL storage path")
	dataBase := flag.String("d", "", "Database connection address")
	httpsEnabled := flag.Bool("s", false, "Enable HTTPS")
	flag.Parse()

	// Переопределение значений из переменных окружения, если они заданы.
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		*addr = envAddr
	}
	if envBase := os.Getenv("BASE_URL"); envBase != "" {
		*base = envBase
	}
	if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
		*storagePath = envStoragePath
	}
	if envDataBase := os.Getenv("DATABASE_DSN"); envDataBase != "" {
		*dataBase = envDataBase
	}
	if envHTTPSEnabled := os.Getenv("ENABLE_HTTPS"); envHTTPSEnabled == "true" {
		*httpsEnabled = true
	}

	// Возвращает инициализированную структуру Flags.
	return &Flags{
		RunAddr:            *addr,
		BaseShortAddr:      *base,
		URLStorageFilePath: *storagePath,
		DataBaseDSN:        *dataBase,
		HTTPSEnabled:       *httpsEnabled,
	}
}
