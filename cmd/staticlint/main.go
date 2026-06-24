package main

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/gostaticanalysis/nilerr"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// OsExitAnalyzer — собственный анализатор, проверяющий запрет на вызов os.Exit
// внутри функции main() пакета main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "checks for direct calls to os.Exit in main function of main package",
	Run:  runOsExitAnalyzer,
}

func main() {
	var mychecks []*analysis.Analyzer

	mychecks = append(mychecks,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpcallCheck(), // Обертка над httpresponse
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// S1000: Use channel select instead of single-case select
	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1000" {
			mychecks = append(mychecks, v.Analyzer)
			break
		}
	}
	// ST1000: Incorrect package comment format
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" {
			mychecks = append(mychecks, v.Analyzer)
			break
		}
	}
	// QF1001: Suggests using strings.Contains
	for _, v := range quickfix.Analyzers {
		if v.Analyzer.Name == "QF1001" {
			mychecks = append(mychecks, v.Analyzer)
			break
		}
	}

	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, nilerr.Analyzer)

	mychecks = append(mychecks, OsExitAnalyzer)

	multichecker.Main(mychecks...)
}

// httpcallCheck возвращает стандартный анализатор httpresponse
func httpcallCheck() *analysis.Analyzer {
	return httpcallCheckWithAlias(httpresponse.Analyzer)
}

func httpcallCheckWithAlias(a *analysis.Analyzer) *analysis.Analyzer {
	return a
}

func runOsExitAnalyzer(pass *analysis.Pass) (interface{}, error) {
	// Проверяем только пакет main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Исключаем тестовые пакеты (Go компилирует тесты во временные пакеты с суффиксом .test)
	if strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil
	}

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename

		// Исключаем файлы из кэша сборки Go и файлы автогенерации тестов
		if strings.Contains(filename, "go-build") || strings.HasSuffix(filename, "_testmain.go") {
			continue
		}

		// Дополнительная проверка: пропускаем файлы с комментариями автогенерации
		if isGeneratedFile(file) {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			// Ищем объявления функций
			fn, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Проверяем, что это функция main
			if fn.Name.Name != "main" {
				return true
			}

			// Обходим AST внутри тела функции main
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// Проверяем левую часть выражения (селектор пакета)
				if pkgIdent, ok := sel.X.(*ast.Ident); ok {
					if obj, ok := pass.TypesInfo.Uses[pkgIdent]; ok {
						if pkgName, ok := obj.(*types.PkgName); ok && pkgName.Imported().Path() == "os" {
							// Проверяем правую часть выражения (имя вызываемой функции)
							if sel.Sel.Name == "Exit" {
								pass.Reportf(call.Pos(), "direct call to os.Exit is forbidden in main function of main package")
							}
						}
					}
				}
				return true
			})

			// Возвращаем false, чтобы не обходить внутренности этой функции повторно
			return false
		})
	}

	return nil, nil
}

// isGeneratedFile проверяет, является ли файл автогенерируемым (например, 'go test')
func isGeneratedFile(file *ast.File) bool {
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if strings.Contains(comment.Text, "Code generated by") {
				return true
			}
		}
	}
	return false
}
