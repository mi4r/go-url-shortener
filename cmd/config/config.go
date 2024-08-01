package config

import (
	"flag"
	"fmt"
	"os"
)

type Flags struct {
	RunAddr            string
	BaseShortAddr      string
	URLStorageFilePath string
}

func (f *Flags) String() string {
	return fmt.Sprintf("RunAddr: %s, BaseShortAddr: %s, URLStorageFileName: %s", f.RunAddr, f.BaseShortAddr, f.URLStorageFilePath)
}

func Init() *Flags {
	addr := flag.String("a", "localhost:8080", "Address and port to run server")
	base := flag.String("b", "http://localhost:8080", "Base shorten url")
	storagePath := flag.String("f", "url_storage.json", "URL storage path")
	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		*addr = envAddr
	}
	if envBase := os.Getenv("BASE_URL"); envBase != "" {
		*base = envBase
	}
	if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
		*storagePath = envStoragePath
	}

	return &Flags{
		RunAddr:            *addr,
		BaseShortAddr:      *base,
		URLStorageFilePath: *storagePath,
	}
}
