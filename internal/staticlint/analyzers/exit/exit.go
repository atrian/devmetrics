package exit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check main.go for prohibited os.Exit calls",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// проходим по всем файлам
	for _, file := range pass.Files {

		// обрабатываем только main
		if file.Name.Name != "main" {
			continue
		}

		// проходим по AST файла
		ast.Inspect(file, func(n ast.Node) bool {
			// проверяем явзяется ли нода декларацией функции
			funcCandidate, isFunc := n.(*ast.FuncDecl)

			// проверяем имя функции
			if !isFunc || funcCandidate.Name.Name != "main" {
				return true
			}

			// проверяем функцию main
			ast.Inspect(n, func(node ast.Node) bool {
				if funcCallCandidate, isCallExpr := node.(*ast.CallExpr); isCallExpr {
					if fun, ok := funcCallCandidate.Fun.(*ast.SelectorExpr); ok {
						if fun.Sel.Name == "Exit" {
							pass.Reportf(funcCallCandidate.Pos(), "os.Exit calls is prohibited in main.go")
						}
					}
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
