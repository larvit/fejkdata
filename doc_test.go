package fejkdata

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"testing"
)

var backticked = regexp.MustCompile("`([A-Za-z][A-Za-z0-9]*(?:\\.[A-Za-z][A-Za-z0-9]*)?)`")

// TestVocabularyNamesDeclaredSymbols proves every symbol the package doc's Vocabulary
// names is declared here, so a renamed unit renames its entry with it.
func TestVocabularyNamesDeclaredSymbols(t *testing.T) {
	files := sourceFiles(t)
	declared := declaredSymbols(files)
	for _, m := range backticked.FindAllStringSubmatch(vocabulary(t, files), -1) {
		if !declared[m[1]] {
			t.Errorf("the Vocabulary names %s, which the package declares nowhere", m[1])
		}
	}
}

// sourceFiles parses the package's own files, the tests left out: the Vocabulary
// indexes what ships, not what proves it.
func sourceFiles(t *testing.T) []*ast.File {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read the package directory: %v", err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		files = append(files, f)
	}
	return files
}

// vocabulary is the package doc from its Vocabulary heading on.
func vocabulary(t *testing.T, files []*ast.File) string {
	t.Helper()
	for _, f := range files {
		if f.Doc == nil {
			continue
		}
		if _, section, found := strings.Cut(f.Doc.Text(), "\n# Vocabulary\n"); found {
			return section
		}
	}
	t.Fatal("no package doc carries a Vocabulary heading, so fence, hold, pin and the rest are defined nowhere a reader of the code meets them")
	return ""
}

// declaredSymbols is every name the package declares, a method and a struct field
// under its type as well: fence, draws.pin, tableRead.whole.
func declaredSymbols(files []*ast.File) map[string]bool {
	names := map[string]bool{}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.FuncDecl:
				names[d.Name.Name] = true
				if d.Recv != nil {
					names[receiverType(d.Recv)+"."+d.Name.Name] = true
				}
			case *ast.TypeSpec:
				names[d.Name.Name] = true
				addFields(names, d)
			case *ast.ValueSpec:
				for _, id := range d.Names {
					names[id.Name] = true
				}
			}
			return true
		})
	}
	return names
}

func addFields(names map[string]bool, d *ast.TypeSpec) {
	s, isStruct := d.Type.(*ast.StructType)
	if !isStruct {
		return
	}
	for _, field := range s.Fields.List {
		for _, id := range field.Names {
			names[d.Name.Name+"."+id.Name] = true
		}
	}
}

func receiverType(recv *ast.FieldList) string {
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
