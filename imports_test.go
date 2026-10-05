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
		".":                  {"internal/drawstate"},
		"cmd/fejkdata":       {"."},
		"internal/drawstate": nil,
	}
	seen := map[string]bool{}
	err = filepath.WalkDir(".", func(dir string, d fs.DirEntry, err error) error {
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
		dir = filepath.ToSlash(dir)
		seen[dir] = true
		may, listed := allowed[dir]
		if !listed {
			t.Errorf("package %s is not listed: add the module packages it may import", dir)
			return nil
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
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatalf("no package imports a package of module %s, which go.mod names", module)
	}
	for dir := range allowed {
		if !seen[dir] {
			t.Errorf("listed package %s does not exist", dir)
		}
	}
}
