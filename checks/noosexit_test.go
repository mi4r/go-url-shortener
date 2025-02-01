package checks

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoOsExitInMainAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	// Запускаем тесты для анализатора
	analysistest.Run(t, testdata, NoOsExitInMainAnalyzer, "b", "c")
}
