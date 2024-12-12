package config

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestInit_EnvVariables(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("SERVER_ADDRESS", "0.0.0.0:8081")
	os.Setenv("BASE_URL", "http://env.url")
	os.Setenv("FILE_STORAGE_PATH", "/env/storage")
	os.Setenv("DATABASE_DSN", "mysql://env_user:env_pass@env_host/env_db")

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
	}()

	expected := &Flags{
		RunAddr:            "0.0.0.0:8081",
		BaseShortAddr:      "http://env.url",
		URLStorageFilePath: "/env/storage",
		DataBaseDSN:        "mysql://env_user:env_pass@env_host/env_db",
	}

	actual := Init()

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %+v, but got %+v", expected, actual)
	}
	_ = fmt.Sprint(actual)
}
