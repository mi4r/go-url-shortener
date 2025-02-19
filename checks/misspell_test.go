package checks

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMisspellAnalyzer(t *testing.T) {
	// Указываем директорию с тестовыми файлами
	testdata := analysistest.TestData()
	// Запускаем тесты
	analysistest.Run(t, testdata, NewMisspellAnalyzer(), "a")
}
