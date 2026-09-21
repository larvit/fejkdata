package fejkdata

import (
	"go/ast"
	"sort"
	"testing"
)

// TestNoFunctionSpellsAMethod holds a call to one unit. A method is written with
// its receiver and a function bare, so one name on both reads alike at a call site
// and greps as one unit; two types may share a method name, since the receiver
// standing before it says which.
func TestNoFunctionSpellsAMethod(t *testing.T) {
	_, files := sourceFiles(t)
	funcs, methods := declaredCalls(files)
	names := make([]string, 0, len(funcs))
	for name := range funcs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if recv := methods[name]; recv != "" {
			t.Errorf("func %s and %s.%s are two units under one name; name them apart", name, recv, name)
		}
	}
}

// declaredCalls is every name the package declares as a function, and every name
// it declares as a method with one type declaring it.
func declaredCalls(files []*ast.File) (funcs map[string]bool, methods map[string]string) {
	funcs, methods = map[string]bool{}, map[string]string{}
	for _, f := range files {
		for _, decl := range f.Decls {
			d, isFunc := decl.(*ast.FuncDecl)
			if !isFunc {
				continue
			}
			if recv := receiverType(d.Recv); recv != "" {
				methods[d.Name.Name] = recv
				continue
			}
			funcs[d.Name.Name] = true
		}
	}
	return funcs, methods
}
