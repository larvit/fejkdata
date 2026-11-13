package fejkdata

import (
	"errors"
	"go/build"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var allowedImports = map[string][]string{
	".":                      {"internal/builtinfunc", "internal/datafiles", "internal/datatype", "internal/drawstate", "internal/engine", "internal/engine/tree", "internal/grammar", "internal/invariant", "internal/jsonvalue"},
	"cli":                    {"."},
	"cmd/fejkdata":           {"cli", "data"},
	"internal/avoid":         nil,
	"internal/builtinfunc":   {"internal/avoid", "internal/datatype", "internal/drawstate", "internal/grammar", "internal/invariant", "internal/proven"},
	"internal/datafiles":     {"internal/grammar", "internal/jsonvalue"},
	"internal/datatype":      nil,
	"internal/drawstate":     nil,
	"internal/engine":        {"internal/datafiles", "internal/datatype", "internal/drawstate", "internal/engine/fence", "internal/engine/load", "internal/engine/render", "internal/engine/tree", "internal/grammar", "internal/invariant", "internal/jsonvalue"},
	"internal/engine/fence":  {"internal/builtinfunc", "internal/datatype", "internal/engine/load", "internal/engine/tree", "internal/grammar", "internal/invariant", "internal/proven"},
	"internal/engine/load":   {"internal/avoid", "internal/builtinfunc", "internal/datafiles", "internal/datatype", "internal/engine/tree", "internal/grammar", "internal/invariant", "internal/jsonvalue", "internal/proven", "internal/rows"},
	"internal/engine/render": {"internal/avoid", "internal/drawstate", "internal/engine/tree", "internal/grammar", "internal/invariant"},
	"internal/engine/tree":   {"internal/avoid", "internal/builtinfunc", "internal/datafiles", "internal/datatype", "internal/grammar", "internal/invariant", "internal/rows"},
	"internal/grammar":       nil,
	"internal/invariant":     nil,
	"internal/jsonvalue":     nil,
	"internal/proven":        {"internal/datatype", "internal/grammar", "internal/invariant"},
	"internal/rows":          {"internal/drawstate", "internal/grammar", "internal/invariant"},
}

var allowedTestImports = map[string][]string{
	".":                     {"internal/engine/fence", "internal/engine/load", "internal/engine/render"},
	"cli":                   {"internal/grammar"},
	"internal/engine/fence": {"internal/jsonvalue"},
}

func moduleNamedInGoMod(t *testing.T) string {
	t.Helper()
	mod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(mod), "\n") {
		if f := strings.Fields(line); len(f) > 1 && f[0] == "module" {
			return strings.Trim(f[1], `"`)
		}
	}
	t.Fatal("go.mod names no module")
	return ""
}

type moduleImport struct {
	dir, rel string
	testOnly bool
}

func (imp moduleImport) allowed() bool {
	may := allowedImports[imp.dir]
	if imp.testOnly {
		may = slices.Concat(may, allowedTestImports[imp.dir])
	}
	return imp.rel == imp.dir || slices.Contains(may, imp.rel) || mayImportShippedData(imp.dir, imp.rel, imp.testOnly)
}

func moduleImports(t *testing.T) []moduleImport {
	t.Helper()
	module := moduleNamedInGoMod(t)
	var imports []moduleImport
	eachPackage(t, func(dir string, pkg *build.Package) {
		for _, imp := range slices.Concat(pkg.Imports, pkg.TestImports, pkg.XTestImports) {
			if imp != module && !strings.HasPrefix(imp, module+"/") {
				continue
			}
			rel := strings.TrimPrefix(strings.TrimPrefix(imp, module), "/")
			if rel == "" {
				rel = "."
			}
			imports = append(imports, moduleImport{dir, rel, !slices.Contains(pkg.Imports, imp)})
		}
	})
	if len(imports) == 0 {
		t.Fatalf("no package imports a package of module %s, which go.mod names", module)
	}
	return imports
}

// TestPackageImports holds each package in the module to the module packages it may import.
func TestPackageImports(t *testing.T) {
	for _, imp := range moduleImports(t) {
		t.Run(imp.dir+" imports "+imp.rel, func(t *testing.T) {
			if !imp.allowed() {
				t.Errorf("package %s imports %s, which its list does not allow", imp.dir, imp.rel)
			}
		})
	}
}

func TestEveryPackageIsListedWithTheModulePackagesItMayImport(t *testing.T) {
	eachPackage(t, func(dir string, _ *build.Package) {
		if _, listed := allowedImports[dir]; !listed && !isShippedData(dir) {
			t.Errorf("package %s is not listed: add the module packages it may import", dir)
		}
	})
}

func TestEveryListedPackageExists(t *testing.T) {
	seen := map[string]bool{}
	eachPackage(t, func(dir string, _ *build.Package) { seen[dir] = true })
	for dir := range allowedImports {
		t.Run(dir, func(t *testing.T) {
			if !seen[dir] {
				t.Errorf("listed package %s does not exist", dir)
			}
		})
	}
}

func isShippedData(dir string) bool { return dir == "data" || strings.HasPrefix(dir, "data/") }

// mayImportShippedData reports whether dir may import rel, a package under data/. The data package
// imports them all, a locale data/<ll_CC> only data/geo/<CC>, and the root's tests any. misc and a
// geo package import none.
func mayImportShippedData(dir, rel string, testOnly bool) bool {
	switch {
	case !isShippedData(rel):
		return false
	case dir == "data":
		return true
	case dir == ".":
		return testOnly
	}
	_, country, isLocale := strings.Cut(strings.TrimPrefix(dir, "data/"), "_")
	return isShippedData(dir) && isLocale && rel == "data/geo/"+country
}

// eachPackage calls fn with every Go package in the module and its slash-separated directory.
func eachPackage(t *testing.T, fn func(dir string, pkg *build.Package)) {
	t.Helper()
	err := filepath.WalkDir(".", func(dir string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		if dir != "." && (strings.HasPrefix(d.Name(), ".") || d.Name() == "testdata") {
			return filepath.SkipDir
		}
		pkg, err := build.ImportDir(dir, 0)
		var noGo *build.NoGoError
		if errors.As(err, &noGo) {
			return nil
		}
		if err != nil {
			return err
		}
		fn(filepath.ToSlash(dir), pkg)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
