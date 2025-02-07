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
	RunAddr            string `json:"server_address"`    // Адрес и порт для запуска сервера.
	BaseShortAddr      string `json:"base_url"`          // Базовый URL для сокращенных ссылок.
	URLStorageFilePath string `json:"file_storage_path"` // Путь к файлу для хранения URL (если используется файловое хранилище).
	DataBaseDSN        string `json:"database_dsn"`      // DSN (Data Source Name) для подключения к базе данных.
	HTTPSEnabled       bool   `json:"enable_https"`      // Возможность подключения к HTTPS-серверу
}

// String возвращает строковое представление текущих параметров конфигурации.
func (f *Flags) String() string {
	return fmt.Sprintf("RunAddr: %s, BaseShortAddr: %s, URLStorageFileName: %s, DataBaseDSN: %s",
		f.RunAddr, f.BaseShortAddr, f.URLStorageFilePath, f.DataBaseDSN)
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
	// configFile := flag.String("c", "", "Path to JSON config file")
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
	// if envConfigFile := os.Getenv("CONFIG"); envConfigFile != "" {
	// 	*configFile = envConfigFile
	// }

	config := Flags{
		RunAddr:            *addr,
		BaseShortAddr:      *base,
		URLStorageFilePath: *storagePath,
		DataBaseDSN:        *dataBase,
		HTTPSEnabled:       *httpsEnabled,
	}

	// if *configFile != "" {
	// 	file, err := os.Open(*configFile)
	// 	if err == nil {
	// 		defer file.Close()
	// 		decoder := json.NewDecoder(file)
	// 		var fileConfig Flags
	// 		if err := decoder.Decode(&fileConfig); err == nil {
	// 			if *addr == "localhost:8080" && fileConfig.RunAddr != "" {
	// 				config.RunAddr = fileConfig.RunAddr
	// 			}
	// 			if *base == "http://localhost:8080" && fileConfig.BaseShortAddr != "" {
	// 				config.BaseShortAddr = fileConfig.BaseShortAddr
	// 			}
	// 			if *storagePath == "" && fileConfig.URLStorageFilePath != "" {
	// 				config.URLStorageFilePath = fileConfig.URLStorageFilePath
	// 			}
	// 			if *dataBase == "" && fileConfig.DataBaseDSN != "" {
	// 				config.DataBaseDSN = fileConfig.DataBaseDSN
	// 			}
	// 			if !*httpsEnabled && fileConfig.HTTPSEnabled {
	// 				config.HTTPSEnabled = true
	// 			}
	// 		}
	// 	}
	// }

	return &config
}
