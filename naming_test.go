package fejkdata

import (
	"go/ast"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestNoFunctionSpellsAMethod(t *testing.T) {
	funcs, methods := declaredCalls(namespaceFiles(t))
	names := make([]string, 0, len(funcs))
	for name := range funcs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if recv := methods[name]; len(recv) > 0 {
			sort.Strings(recv)
			t.Errorf("func %s is spelled by %s too; name them apart", name, strings.Join(recv, " and "))
		}
	}
}

func namespaceFiles(t *testing.T) []*ast.File {
	t.Helper()
	pkg := packageDir(t)
	_, files := parseFiles(t, slices.Concat(pkg.GoFiles, pkg.TestGoFiles))
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
