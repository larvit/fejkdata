package fejkdata

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"sort"
	"strings"
	"testing"
)

// docs/decisions.md#one-name-one-meaning
func TestNoFunctionSpellsAMethod(t *testing.T) {
	funcs, methods := declaredCalls(namespaceFiles(t))
	for _, name := range sortedNames(funcs) {
		if recv := methods[name]; len(recv) > 0 {
			sort.Strings(recv)
			t.Errorf("func %s is spelled by %s too; name them apart", name, strings.Join(recv, " and "))
		}
	}
}

func namespaceFiles(t *testing.T) []*ast.File {
	t.Helper()
	pkg := packageDir(t)
	fset := token.NewFileSet()
	var files []*ast.File
	for _, name := range slices.Concat(pkg.GoFiles, pkg.TestGoFiles) {
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		files = append(files, f)
	}
	return files
}

func declaredCalls(files []*ast.File) (funcs map[string]bool, methods map[string][]string) {
	funcs, methods = map[string]bool{}, map[string][]string{}
	for _, f := range files {
		for _, decl := range f.Decls {
			d, isFunc := decl.(*ast.FuncDecl)
			if !isFunc {
				continue
			}
			if recv := receiverType(d.Recv); recv != "" {
				methods[d.Name.Name] = append(methods[d.Name.Name], recv+"."+d.Name.Name)
				continue
			}
			funcs[d.Name.Name] = true
		}
	}
	return funcs, methods
}

func receiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	e := recv.List[0].Type
	if star, isPointer := e.(*ast.StarExpr); isPointer {
		e = star.X
	}
	id, isIdent := e.(*ast.Ident)
	if !isIdent {
		return ""
	}
	return id.Name
}
