// Package osexit предоставляет статический анализатор, запрещающий прямые
// вызовы os.Exit внутри функции main пакета main.
package osexit

import (
	"go/ast"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"
)

// generatedRe соответствует каноническому заголовку "Code generated ... DO NOT EDIT.",
// описанному в https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source
var generatedRe = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

// Analyzer сообщает о прямых вызовах os.Exit внутри функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "запрещает прямые вызовы os.Exit внутри функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	for _, file := range pass.Files {
		if isGenerated(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if isOSExitCall(pass, call) {
					pass.Reportf(call.Pos(), "прямой вызов os.Exit в main.main запрещён")
				}
				return true
			})
		}
	}
	return nil, nil
}

func isOSExitCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Exit" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return false
	}
	return pkgName.Imported().Path() == "os"
}

func isGenerated(_ *analysis.Pass, file *ast.File) bool {
	for _, cg := range file.Comments {
		if cg.Pos() >= file.Package {
			break
		}
		for _, c := range cg.List {
			if generatedRe.MatchString(c.Text) {
				return true
			}
		}
	}
	return false
}
