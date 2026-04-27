package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

const marker = "// generate:reset"

// generatedRe соответствует каноническому заголовку "Code generated ... DO NOT EDIT.",
// описанному в https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source
var generatedRe = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

type targetStruct struct {
	pkg        *packages.Package
	file       *ast.File
	typeSpec   *ast.TypeSpec
	structType *ast.StructType
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	pkgs, err := loadPackages(dir)
	if err != nil {
		log.Fatalf("ошибка загрузки package: %v", err)
	}

	targets := discover(pkgs)
	if len(targets) == 0 {
		return
	}
	for _, t := range targets {
		pos := t.pkg.Fset.Position(t.typeSpec.Pos())
		fmt.Fprintf(os.Stderr, "found %s.%s\t%s:%d\n", t.pkg.PkgPath, t.typeSpec.Name.Name, pos.Filename, pos.Line)
	}
	if err := generate(targets); err != nil {
		log.Fatalf("ошибка генерации: %v", err)
	}
}

func loadPackages(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:   dir,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	var hasErrors bool
	for _, p := range pkgs {
		for _, e := range p.Errors {
			fmt.Fprintln(os.Stderr, e)
			hasErrors = true
		}
	}
	if hasErrors {
		return nil, fmt.Errorf("ошибка компиляции в коде")
	}
	return pkgs, nil
}

func discover(pkgs []*packages.Package) []targetStruct {
	var out []targetStruct
	for _, p := range pkgs {
		for _, f := range p.Syntax {
			if isGenerated(f) {
				continue
			}
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				groupDoc := gen.Doc
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					doc := ts.Doc
					if doc == nil {
						doc = groupDoc
					}
					if !hasMarker(doc) {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						fmt.Fprintf(os.Stderr, "внимание: %s помечен, но не является структурой\n", ts.Name.Name)
						continue
					}
					out = append(out, targetStruct{
						pkg:        p,
						file:       f,
						typeSpec:   ts,
						structType: st,
					})
				}
			}
		}
	}
	return out
}

func isGenerated(f *ast.File) bool {
	for _, cg := range f.Comments {
		// заголовок DO NOT EDIT должен находиться до package-объявления
		if cg.Pos() >= f.Package {
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

func hasMarker(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if strings.TrimSpace(c.Text) == marker {
			return true
		}
	}
	return false
}
