package fejkdata

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// The release tags every tracked go.mod's module, so the gate covers every go.mod.
func TestEveryGoModIsInGoWork(t *testing.T) {
	var used []string
	for _, m := range workspaceModules(t) {
		used = append(used, m.dir)
	}
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "." && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if dir := filepath.ToSlash(filepath.Dir(path)); d.Name() == "go.mod" && !slices.Contains(used, dir) {
			t.Errorf("%s/go.mod is not in go.work's use list, so the gate does not cover it", dir)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoGoModRequiresAWorkspaceModule(t *testing.T) {
	mods := workspaceModules(t)
	paths := map[string]bool{}
	for _, m := range mods {
		paths[m.Module.Path] = true
	}
	for _, m := range mods {
		for _, r := range m.Require {
			if paths[r.Path] {
				t.Errorf("%s/go.mod requires %s: go.work joins the modules, and the release commit writes their requires", m.dir, r.Path)
			}
		}
	}
}

func TestNoGoModCarriesReplace(t *testing.T) {
	for _, m := range workspaceModules(t) {
		for _, r := range m.Replace {
			t.Errorf("%s/go.mod replaces %s: go.work joins the modules, and go install refuses a module carrying a replace", m.dir, r.Old.Path)
		}
	}
}

func TestWorkspaceModulesShareTheLowestGo(t *testing.T) {
	var work struct{ Go string }
	if err := json.Unmarshal(goCmd(t, ".", "work", "edit", "-json"), &work); err != nil {
		t.Fatal(err)
	}
	for _, m := range workspaceModules(t) {
		if m.Go != work.Go {
			t.Errorf("%s/go.mod states go %s, go.work %s: raising the lowest Go raises it in every module", m.dir, m.Go, work.Go)
		}
	}
}

func TestTopChangelogHeadingIsUnreleasedOrAVersion(t *testing.T) {
	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	top := regexp.MustCompile(`(?m)^## \[([^\]]*)\]`).FindSubmatch(changelog)
	if top == nil || !regexp.MustCompile(`^(Unreleased|\d+\.\d+\.\d+)$`).Match(top[1]) {
		t.Errorf("CHANGELOG.md's top heading is not [Unreleased] or [X.Y.Z], which the release job refuses")
	}
}

type modulePath struct{ Path, Version string }

type goMod struct {
	Module  modulePath
	Go      string
	Require []modulePath
	Replace []struct{ Old modulePath }

	dir string
}

// workspaceModules returns the modules go.work uses, as the go command reads them.
func workspaceModules(t *testing.T) []goMod {
	t.Helper()
	var work struct{ Use []struct{ DiskPath string } }
	if err := json.Unmarshal(goCmd(t, ".", "work", "edit", "-json"), &work); err != nil {
		t.Fatal(err)
	}
	var mods []goMod
	for _, u := range work.Use {
		m := goMod{dir: filepath.ToSlash(filepath.Clean(u.DiskPath))}
		if err := json.Unmarshal(goCmd(t, m.dir, "mod", "edit", "-json"), &m); err != nil {
			t.Fatal(err)
		}
		mods = append(mods, m)
	}
	return mods
}

func goCmd(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if exit := (*exec.ExitError)(nil); errors.As(err, &exit) {
			stderr = exit.Stderr
		}
		t.Fatalf("go %v in %s: %v\n%s", args, dir, err, stderr)
	}
	return out
}
