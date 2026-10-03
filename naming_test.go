package fejkdata

import (
	"go/ast"
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

func TestEveryPanicIsAnInternalError(t *testing.T) {
	fset, files := sourceFiles(t)
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall || !isIdent(call.Fun, "panic") {
				return true
			}
			if arg, isArgCall := call.Args[0].(*ast.CallExpr); !isArgCall || !isIdent(arg.Fun, "internalError") {
				t.Errorf("%s: panic with internalError(…), so every invariant break greps to one phrase", fset.Position(call.Pos()))
			}
			return true
		})
	}
}

func isIdent(e ast.Expr, name string) bool {
	id, isID := e.(*ast.Ident)
	return isID && id.Name == name
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
