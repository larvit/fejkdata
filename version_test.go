package fejkdata

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
)

const versionFile = "version.go"

// REGENERATE=1 rewrites Version and every workspace module's require of another.
func TestVersionIsTheNewestChangelogHeading(t *testing.T) {
	want := newestRelease(t)
	if os.Getenv("REGENERATE") == "1" {
		src, err := os.ReadFile(versionFile)
		if err != nil {
			t.Fatal(err)
		}
		src = regexp.MustCompile(`(?m)^const Version = ".*"$`).ReplaceAll(src, []byte(`const Version = "`+want+`"`))
		if err := os.WriteFile(versionFile, src, 0o644); err != nil {
			t.Fatal(err)
		}
		for _, m := range workspaceModules(t) {
			for _, r := range m.requiresOfWorkspace {
				goCmd(t, m.dir, "mod", "edit", "-require="+r.Path+"@"+want)
			}
		}
		return
	}
	if Version != want {
		t.Errorf("Version = %q, want %q from CHANGELOG.md's newest versioned heading; regenerate: docker compose run --rm --user \"$(id -u):$(id -g)\" generate", Version, want)
	}
}

func TestWorkspaceRequiresAreAtVersion(t *testing.T) {
	for _, m := range workspaceModules(t) {
		for _, r := range m.requiresOfWorkspace {
			if r.Version != Version {
				t.Errorf("%s/go.mod requires %s %s, want %s; regenerate: docker compose run --rm --user \"$(id -u):$(id -g)\" generate", m.dir, r.Path, r.Version, Version)
			}
		}
	}
}

func TestEveryGoModIsInGoWork(t *testing.T) {
	var used []string
	for _, m := range workspaceModules(t) {
		used = append(used, m.dir)
	}
	for _, dir := range sourceDirs(t) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !slices.Contains(used, dir) {
			t.Errorf("%s/go.mod is not in go.work's use list, so the gate does not cover it", dir)
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

func newestRelease(t *testing.T) string {
	t.Helper()
	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	if m := regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\]`).FindSubmatch(changelog); m != nil {
		return "v" + string(m[1])
	}
	return "v0.0.0"
}

type modulePath struct{ Path, Version string }

type goMod struct {
	Module  modulePath
	Require []modulePath
	Replace []struct{ Old modulePath }

	dir                 string
	requiresOfWorkspace []modulePath
}

// workspaceModules returns the modules go.work uses, as the go command reads them.
func workspaceModules(t *testing.T) []goMod {
	t.Helper()
	var work struct{ Use []struct{ DiskPath string } }
	if err := json.Unmarshal(goCmd(t, ".", "work", "edit", "-json"), &work); err != nil {
		t.Fatal(err)
	}
	var mods []goMod
	paths := map[string]bool{}
	for _, u := range work.Use {
		var m goMod
		m.dir = filepath.ToSlash(filepath.Clean(u.DiskPath))
		if err := json.Unmarshal(goCmd(t, m.dir, "mod", "edit", "-json"), &m); err != nil {
			t.Fatal(err)
		}
		mods = append(mods, m)
		paths[m.Module.Path] = true
	}
	for i, m := range mods {
		for _, r := range m.Require {
			if paths[r.Path] {
				mods[i].requiresOfWorkspace = append(mods[i].requiresOfWorkspace, r)
			}
		}
	}
	return mods
}

func goCmd(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go %v in %s: %v", args, dir, err)
	}
	return out
}
