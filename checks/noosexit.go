package checks

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitInMainAnalyzer — анализатор, запрещающий вызов os.Exit в функции main пакета main.
var NoOsExitInMainAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "запрещает вызов os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		// Игнорируем файлы из кэша
		if strings.Contains(filename, "/go-build/") {
			continue
		}

		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				x, ok := sel.X.(*ast.Ident)
				if !ok || x.Name != "os" {
					return true
				}

				if sel.Sel.Name == "Exit" {
					fmt.Printf("Checking node: %T at position: %v\n", node, pass.Fset.Position(node.Pos()))
					pass.Reportf(call.Pos(), "direct call to os.Exit is not allowed in main function")
				}

				return true
			})
			return false
		})
	}
	return nil, nil
}
