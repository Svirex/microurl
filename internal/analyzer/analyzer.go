package analyzer

import (
	"fmt"
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
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.FuncDecl:
					if x.Name.Name == "main" {
						printMainCallExpr(pass, x.Body.List)
					}
				}
				return true
			})
		}

	}
	return nil, nil
}

func printMainCallExpr(pass *analysis.Pass, stmts []ast.Stmt) {
	for _, stmt := range stmts {
		switch x := stmt.(type) {
		case *ast.ExprStmt:
			if call, ok := x.X.(*ast.CallExpr); ok {
				fmt.Println(call.Fun)
				// fmt.Println(pass.TypesInfo.Types[call.Fun].Type.String())

				switch t := call.Fun.(type) {
				case *ast.SelectorExpr:
					// fmt.Println(t.Sel.Name)
					switch d := t.X.(type) {
					case *ast.Ident:
						fmt.Println(d, d.Name, d.Pos())
					}
				}
			}
		case *ast.DeclStmt: 
			fmt.