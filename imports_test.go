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

// TestPackageImports holds each package in the repository to the module packages it may import.
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
		".":                    {"internal/builtinfunc", "internal/datafiles", "internal/datatype", "internal/drawstate", "internal/grammar", "internal/invariant", "internal/jsonvalue", "internal/proven", "internal/rows"},
		"cmd/fejkdata":         {"."},
		"internal/builtinfunc": {"internal/datatype", "internal/drawstate", "internal/grammar", "internal/invariant", "internal/proven"},
		"internal/datafiles":   {"internal/grammar", "internal/jsonvalue"},
		"internal/datatype":    nil,
		"internal/drawstate":   nil,
		"internal/grammar":     nil,
		"internal/invariant":   nil,
		"internal/jsonvalue":   nil,
		"internal/proven":      {"internal/datatype", "internal/grammar", "internal/invariant"},
		"internal/rows":        {"internal/drawstate", "internal/grammar", "internal/invariant"},
	}
	testsAlso := map[string][]string{
		"cmd/fejkdata": {"internal/grammar"},
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
			may := may
			if !slices.Contains(pkg.Imports, imp) {
				may = slices.Concat(may, testsAlso[dir])
			}
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

// eachPackage calls fn with every Go package in the repository and its slash-separated directory.
func eachPackage(t *testing.T, fn func(dir string, pkg *build.Package)) {
	t.Helper()
	for _, dir := range sourceDirs(t) {
		pkg, err := build.ImportDir(dir, 0)
		var noGo *build.NoGoError
		if errors.As(err, &noGo) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		fn(dir, pkg)
	}
}

// sourceDirs returns every slash-separated directory of the repository but hidden ones and testdata.
func sourceDirs(t *testing.T) []string {
	t.Helper()
	var dirs []string
	err := filepath.WalkDir(".", func(dir string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		if dir != "." && (strings.HasPrefix(d.Name(), ".") || d.Name() == "testdata") {
			return filepath.SkipDir
		}
		dirs = append(dirs, filepath.ToSlash(dir))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return dirs
}
