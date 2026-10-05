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

// TestPackageImports holds each package in the module to the module packages it may import.
func TestPackageImports(t *testing.T) {
	mod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	module := ""
	for _, line := range strings.Split(string(mod), "\n") {
		if f := strings.Fields(line); len(f) > 1 && f[0] == "module" {
			module = strings.Trim(f[1], `"`)
			break
		}
	}
	if module == "" {
		t.Fatal("go.mod names no module")
	}
	checked := 0
	allowed := map[string][]string{
		".":                    {"internal/builtinfunc", "internal/datafiles", "internal/datatype", "internal/drawstate", "internal/grammar", "internal/invariant", "internal/proven"},
		"cmd/fejkdata":         {"."},
		"internal/builtinfunc": {"internal/datatype", "internal/drawstate", "internal/grammar", "internal/invariant", "internal/proven"},
		"internal/datafiles":   {"internal/grammar"},
		"internal/datatype":    nil,
		"internal/drawstate":   nil,
		"internal/grammar":     nil,
		"internal/invariant":   nil,
		"internal/proven":      {"internal/datatype", "internal/grammar", "internal/invariant"},
	}
	seen := map[string]bool{}
	eachPackage(t, func(dir string, pkg *build.Package) {
		seen[dir] = true
		may, listed := allowed[dir]
		if !listed {
			t.Errorf("package %s is not listed: add the module packages it may import", dir)
			return
		}
		for _, imp := range slices.Concat(pkg.Imports, pkg.TestImports, pkg.XTestImports) {
			if imp != module && !strings.HasPrefix(imp, module+"/") {
				continue
			}
			checked++
			rel := strings.TrimPrefix(strings.TrimPrefix(imp, module), "/")
			if rel == "" {
				rel = "."
			}
			if rel != dir && !slices.Contains(may, rel) {
				t.Errorf("package %s imports %s, which its list does not allow", dir, rel)
			}
		}
	})
	if checked == 0 {
		t.Fatalf("no package imports a package of module %s, which go.mod names", module)
	}
	for dir := range allowed {
		if !seen[dir] {
			t.Errorf("listed package %s does not exist", dir)
		}
	}
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
