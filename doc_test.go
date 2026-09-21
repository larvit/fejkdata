package fejkdata

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// Every backtick in the Vocabulary is a symbol.
var backticked = regexp.MustCompile("`([^`\n]+)`")

// TestVocabularyNamesDeclaredSymbols holds a renamed unit to its Vocabulary entry.
func TestVocabularyNamesDeclaredSymbols(t *testing.T) {
	_, files := sourceFiles(t)
	declared := declaredSymbols(files)
	if len(namesDeclared(t, "the Vocabulary", vocabulary(t, files), declared)) == 0 {
		t.Fatal("the Vocabulary names no symbol, so nothing holds a renamed unit to its entry")
	}
}

// namesDeclared is every symbol text backticks, failing for each one the package
// declares nowhere.
func namesDeclared(t *testing.T, where, text string, declared map[string]bool) []string {
	t.Helper()
	var named []string
	for _, m := range backticked.FindAllStringSubmatch(text, -1) {
		if !declared[m[1]] {
			t.Errorf("%s names %s, which the package declares nowhere", where, m[1])
		}
		named = append(named, m[1])
	}
	return named
}

// namesAPass fails unless one symbol text backticks is a function: a field names
// what a pass reads or writes, never the pass.
func namesAPass(t *testing.T, where, text string, declared, funcs map[string]bool) {
	t.Helper()
	named := namesDeclared(t, where, text, declared)
	if !slices.ContainsFunc(named, func(n string) bool { return funcs[n] }) {
		t.Errorf("%s names no function, so it names no pass", where)
	}
}

func sourceFiles(t *testing.T) (*token.FileSet, []*ast.File) {
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
	return fset, files
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
	names := declaredFuncs(files)
	for _, f := range files {
		for _, decl := range f.Decls {
			if d, isGen := decl.(*ast.GenDecl); isGen {
				addSpecs(names, d)
			}
		}
	}
	return names
}

// declaredFuncs is every function the package declares, a method keyed under its
// type: template.compileFormat.
func declaredFuncs(files []*ast.File) map[string]bool {
	names := map[string]bool{}
	for _, f := range files {
		for _, decl := range f.Decls {
			d, isFunc := decl.(*ast.FuncDecl)
			if !isFunc {
				continue
			}
			names[d.Name.Name] = true
			if recv := receiverType(d.Recv); recv != "" {
				names[recv+"."+d.Name.Name] = true
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

const fillHeader = "Filled by "

// TestTemplateFieldGroupsNameADeclaredPass holds a renamed pass to the group it
// fills, and to the template doc.
func TestTemplateFieldGroupsNameADeclaredPass(t *testing.T) {
	fset, files := sourceFiles(t)
	declared, funcs := declaredSymbols(files), declaredFuncs(files)
	doc, st, headers := templateStruct(t, fset, files)
	namesAPass(t, "the template doc", doc, declared, funcs)
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			t.Error("an embedded field of template sits under no header, so nothing says which pass fills it")
		}
		for _, id := range field.Names {
			header := headerAbove(headers, id.Pos())
			if header == "" {
				t.Errorf("template.%s sits under no %q header, so nothing says which pass fills it", id.Name, fillHeader)
				continue
			}
			namesAPass(t, "the header above template."+id.Name, header, declared, funcs)
		}
	}
}

// templateStruct is the template struct, its doc, and the fill headers standing
// between its fields, in source order.
func templateStruct(t *testing.T, fset *token.FileSet, files []*ast.File) (string, *ast.StructType, []*ast.CommentGroup) {
	t.Helper()
	for _, f := range files {
		for _, decl := range f.Decls {
			d, isGen := decl.(*ast.GenDecl)
			if !isGen {
				continue
			}
			for _, spec := range d.Specs {
				s, isType := spec.(*ast.TypeSpec)
				if !isType || s.Name.Name != "template" {
					continue
				}
				st, isStruct := s.Type.(*ast.StructType)
				if !isStruct {
					t.Fatal("template is not a struct")
				}
				return d.Doc.Text() + s.Doc.Text(), st, fillHeaders(fset, f, st)
			}
		}
	}
	t.Fatal("the package declares no template struct")
	return "", nil, nil
}

// fillHeaders skips a header trailing a field: it names that field, never the
// fields below it.
func fillHeaders(fset *token.FileSet, f *ast.File, st *ast.StructType) []*ast.CommentGroup {
	trailing := map[int]bool{}
	for _, field := range st.Fields.List {
		trailing[fset.Position(field.End()).Line] = true
	}
	var found []*ast.CommentGroup
	for _, c := range f.Comments {
		if c.Pos() < st.Pos() || c.End() > st.End() || trailing[fset.Position(c.Pos()).Line] {
			continue
		}
		if strings.HasPrefix(c.Text(), fillHeader) {
			found = append(found, c)
		}
	}
	return found
}

// headerAbove is the last fill header standing above pos.
func headerAbove(headers []*ast.CommentGroup, pos token.Pos) string {
	text := ""
	for _, c := range headers {
		if c.Pos() < pos {
			text = c.Text()
		}
	}
	return text
}
