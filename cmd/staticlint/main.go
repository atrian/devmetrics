package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/atrian/devmetrics/internal/staticlint"
)

func main() {
	linter := staticlint.NewCheckerOptionsWithAllRules()

	// запускаем мультичекер со своим набором правил
	multichecker.Main(linter.GetRules()...)
}
