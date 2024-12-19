/*
Package staticlint реализует multichecker для анализа Go-кода.

Этот multichecker включает:
1. Стандартные анализаторы из пакета `golang.org/x/tools/go/analysis/passes`.
2. Анализаторы класса SA из пакета `staticcheck.io`.
3. Один анализатор из других классов `staticcheck.io` (например, ST1000).
4. Два публичных анализатора: ineffassign, misspell.
5. Собственный анализатор `noosexit`, запрещающий вызов os.Exit в функции main пакета main.

Запуск:

	go run ./cmd/staticlint <путь к проекту>

Каждый анализатор проверяет разные аспекты качества и безопасности кода.
*/
package main

import (
	"github.com/mi4r/go-url-shortener/checks"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	// Добавляем стандартные анализаторы
	for _, a := range staticcheck.Analyzers {
		if a.Analyzer != nil {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	// Добавляем анализатор misspell
	analyzers = append(analyzers, checks.NewMisspellAnalyzer())

	analyzers = append(analyzers, checks.NoOsExitInMainAnalyzer)

	multichecker.Main(analyzers...)
}
