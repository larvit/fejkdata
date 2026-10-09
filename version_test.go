package fejkdata

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const setVersion = "docker compose run --rm --user \"$(id -u):$(id -g)\" release-tooling release-tooling/set_version.py"

func TestVersionIsTheNewestChangelogHeading(t *testing.T) {
	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	want := "v0.0.0"
	if m := regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\]`).FindSubmatch(changelog); m != nil {
		want = "v" + string(m[1])
	}
	if Version != want {
		t.Errorf("Version = %q, want %q from CHANGELOG.md's newest versioned heading; run %s", Version, want, setVersion)
	}
}

func TestWorkspaceModulesRequireEachOtherAtVersion(t *testing.T) {
	uses := workspaceModules(t)
	for _, dir := range goModDirs(t) {
		if !slices.Contains(uses, dir) {
			t.Errorf("%s/go.mod is not in go.work's use list, so the gate does not cover it", dir)
		}
	}
	for _, dir := range uses {
		mod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(mod), "\n") {
			f := strings.Fields(line)
			if len(f) > 0 && f[0] == "replace" {
				t.Errorf("%s/go.mod carries %q: go.work joins the modules, and a published go.mod carries no replace", dir, line)
			}
			if len(f) > 0 && f[0] == "require" {
				f = f[1:]
			}
			if len(f) >= 2 && (f[0] == "github.com/larvit/fejkdata" || strings.HasPrefix(f[0], "github.com/larvit/fejkdata/")) && f[1] != Version {
				t.Errorf("%s/go.mod requires %s %s, want %s; run %s", dir, f[0], f[1], Version, setVersion)
			}
		}
	}
}

// workspaceModules returns the directories go.work's use directives name, cleaned and slash-separated.
func workspaceModules(t *testing.T) []string {
	t.Helper()
	work, err := os.ReadFile("go.work")
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	inUse := false
	for _, line := range strings.Split(string(work), "\n") {
		f := strings.Fields(line)
		switch {
		case inUse && len(f) == 1 && f[0] == ")":
			inUse = false
		case inUse && len(f) == 1:
			dirs = append(dirs, filepath.ToSlash(filepath.Clean(f[0])))
		case len(f) == 2 && f[0] == "use" && f[1] == "(":
			inUse = true
		case len(f) == 2 && f[0] == "use":
			dirs = append(dirs, filepath.ToSlash(filepath.Clean(f[1])))
		}
	}
	if len(dirs) == 0 {
		t.Fatal("go.work uses no module")
	}
	return dirs
}

func goModDirs(t *testing.T) []string {
	t.Helper()
	var dirs []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "." && (strings.HasPrefix(d.Name(), ".") || d.Name() == "testdata") {
			return filepath.SkipDir
		}
		if d.Name() == "go.mod" {
			dirs = append(dirs, filepath.ToSlash(filepath.Dir(path)))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return dirs
}
