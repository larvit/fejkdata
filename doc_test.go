package fejkdata

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
	"testing"
)

// Every backtick in the Vocabulary is a symbol.
var backticked = regexp.MustCompile("`([^`\n]+)`")

// TestVocabularyNamesDeclaredSymbols holds a renamed unit to its Vocabulary entry.
func TestVocabularyNamesDeclaredSymbols(t *testing.T) {
	files := sourceFiles(t)
	declared := declaredSymbols(files)
	named := backticked.FindAllStringSubmatch(vocabulary(t, files), -1)
	if len(named) == 0 {
		t.Fatal("the Vocabulary names no symbol, so nothing holds a renamed unit to its entry")
	}
	for _, m := range named {
		if !declared[m[1]] {
			t.Errorf("the Vocabulary names %s, which the package declares nowhere", m[1])
		}
	}
}

func sourceFiles(t *testing.T) []*ast.File {
	t.Helper()
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("read the package directory: %v", err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, name := range pkg.GoFiles {
		f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		files = append(files, f)
	}
	return files
}

func vocabulary(t *testing.T, files []*ast.File) string {
	t.Helper()
	var sections strings.Builder
	for _, f := range files {
		for _, c := range f.Comments {
			section, found := strings.CutPrefix(c.Text(), "Vocabulary\n\n")
			if found && c.Pos() > f.Package {
				sections.WriteString(section)
			}
		}
	}
	if sections.Len() == 0 {
		t.Fatal("no comment below a package clause opens with a Vocabulary heading, so the words the package is written in are defined nowhere a reader of the code meets them")
	}
	return sections.String()
}

// declaredSymbols is every name the package declares at the top level, a method and
// a struct field keyed under its type: draws.pin, tableRead.whole.
func declaredSymbols(files []*ast.File) map[string]bool {
	names := map[string]bool{}
	for _, f := range files {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				names[d.Name.Name] = true
				if recv := receiverType(d.Recv); recv != "" {
					names[recv+"."+d.Name.Name] = true
				}
			case *ast.GenDecl:
				addSpecs(names, d)
			}
		}
	}
	return names
}

func addSpecs(names map[string]bool, d *ast.GenDecl) {
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			names[s.Name.Name] = true
			addFields(names, s)
		case *ast.ValueSpec:
			for _, id := range s.Names {
				names[id.Name] = true
			}
		}
	}
}

func addFields(names map[string]bool, s *ast.TypeSpec) {
	st, isStruct := s.Type.(*ast.StructType)
	if !isStruct {
		return
	}
	for _, field := range st.Fields.List {
		for _, id := range field.Names {
			names[s.Name.Name+"."+id.Name] = true
		}
	}
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
