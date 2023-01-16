// Package staticlint набор статических анализаторов согласно ТЗ:
// стандартные статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
// все анализаторы класса SA пакета staticcheck.io;
// не менее одного анализатора остальных классов пакета staticcheck.io;
// два публичных анализаторов на выбор.
// собственный анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"

	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

// CheckerOptions структура для хранения набора подключенных стат.анализаторов кода
type CheckerOptions struct {
	rules []*analysis.Analyzer
}

// NewCheckerOptions возвращает структура с подготовленным набором анализаторов согласно ТЗ
func NewCheckerOptions() *CheckerOptions {
	co := CheckerOptions{}

	// загружаем все анализаторы
	co.AddGoAnalysisPassesRules()
	co.AddSAStaticCheckIORules()
	co.AddSStaticCheckIORules()
	co.AddQStaticCheckIORules()
	co.AddExternalCheckerRules()
	co.AddCustomCheckerRules()

	return &co
}

// AddCheckerRules добавление анализатора в список проверки утилиты CheckerOptions
func (co *CheckerOptions) AddCheckerRules(rules []*analysis.Analyzer) {
	co.rules = append(co.rules, rules...)
}

// AddGoAnalysisPassesRules добавление стандартных статических анализаторов пакета
// golang.org/x/tools/go/analysis/passes. Подробная информация об анализаторах
// https://pkg.go.dev/golang.org/x/tools/go/analysis/passes
func (co *CheckerOptions) AddGoAnalysisPassesRules() *CheckerOptions {
	co.AddCheckerRules([]*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	})

	return co
}

// AddSAStaticCheckIORules добавление статических анализаторов класса SA пакета
// honnef.co/go/tools/cmd/staticcheck. Подробная информация об анализаторах
// https://staticcheck.io/docs/checks/#SA
func (co *CheckerOptions) AddSAStaticCheckIORules() *CheckerOptions {
	// добавляем новые правила проверки
	co.AddCheckerRules(convertStaticCheckAnalyzer(staticcheck.Analyzers))

	return co
}

// AddSStaticCheckIORules добавление статических анализаторов класса S пакета
// honnef.co/go/tools/cmd/staticcheck. Подробная информация об анализаторах
// https://staticcheck.io/docs/checks/#S
func (co *CheckerOptions) AddSStaticCheckIORules() *CheckerOptions {
	// добавляем новые правила проверки
	co.AddCheckerRules(convertStaticCheckAnalyzer(simple.Analyzers))

	return co
}

// AddQStaticCheckIORules добавление статических анализаторов класса Q пакета
// honnef.co/go/tools/cmd/staticcheck. Подробная информация об анализаторах
// https://staticcheck.io/docs/checks/#S
func (co *CheckerOptions) AddQStaticCheckIORules() *CheckerOptions {
	// добавляем новые правила проверки
	co.AddCheckerRules(convertStaticCheckAnalyzer(quickfix.Analyzers))

	return co
}

func (co *CheckerOptions) AddExternalCheckerRules() *CheckerOptions {
	//TODO любой внешний чекер
	return co
}

func (co *CheckerOptions) AddCustomCheckerRules() *CheckerOptions {
	//TODO свой чекер на os.exit
	return co
}

// convertStaticCheckAnalyzer приводит слайс анализаторов пакета honnef.co/go/tools/cmd/staticcheck
// к общему виду
func convertStaticCheckAnalyzer(analyzers []*lint.Analyzer) []*analysis.Analyzer {
	// резервируем место под анализаторы
	rules := make([]*analysis.Analyzer, 0, len(analyzers))

	// собираем анализаторы в слайс
	for _, rule := range analyzers {
		rules = append(rules, rule.Analyzer)
	}

	return rules
}

// GetRules возвращает слайс со всеми зарегистриррованными анализаторами
func (co *CheckerOptions) GetRules() []*analysis.Analyzer {
	return co.rules
}
