package checks

import (
	"go/token"
	"os"
	"strings"

	"github.com/client9/misspell"
	"golang.org/x/tools/go/analysis"
)

// NewMisspellAnalyzer создает анализатор misspell.
func NewMisspellAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "misspell",
		Doc:  "проверяет орфографию в коде",
		Run: func(pass *analysis.Pass) (interface{}, error) {
			replacer := misspell.New()
			replacer.Compile() // Компиляция словаря исправлений

			for _, file := range pass.Files {
				// Получаем путь к файлу
				filePath := pass.Fset.File(file.Pos()).Name()

				// Читаем содержимое файла
				content, err := os.ReadFile(filePath)
				if err != nil {
					continue // Пропускаем файл, если его нельзя прочитать
				}

				// Проверяем текст на ошибки орфографии
				corrected, _ := replacer.Replace(string(content))
				if corrected != string(content) {
					// Выводим исправления
					lines := strings.Split(string(content), "\n")
					for i, line := range lines {
						if correctedLine, _ := replacer.Replace(line); correctedLine != line {
							pass.Reportf(token.Pos(i+1), "обнаружена ошибка орфографии: %s -> %s", line, correctedLine)
						}
					}
				}
			}
			return nil, nil
		},
	}
}
