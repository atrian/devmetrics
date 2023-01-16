package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/atrian/devmetrics/internal/staticlint"
)

func main() {
	m := staticlint.NewCheckerOptions()

	// запускаем мультичекер со своим набором правил
	multichecker.Main(m.GetRules()...)
}
