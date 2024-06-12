// Пакет analyzer определяет Analyzer, который проверяет, что в функции `main` пакета `main` нет вызова функции `os.Exit`.
package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check unexpected os.Exit in main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				return true
			}
			for _, stmt := range funcDecl.Body.List {
				ast.Inspect(stmt, func(node ast.Node) bool {
					if isOsExitCall(node) {
						pass.Reportf(node.Pos(), "found os.Exit call")
					}
					return true
				})
			}
			return false
		})

	}
	return nil, nil
}

func isOsExitCall(stmt ast.Node) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	callStmt, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selectorExpr, ok := callStmt.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := selectorExpr.Sel.Name
	if name != "Exit" {
		return false
	}

	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok || ident.Name != "os" {
		return false
	}
	return true
}
