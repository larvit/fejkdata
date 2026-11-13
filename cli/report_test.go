package cli

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestVersionLineNamesFejkdatasVersion(t *testing.T) {
	fejkdata := func(version string, replace *debug.Module) *debug.Module {
		return &debug.Module{Path: "github.com/larvit/fejkdata", Version: version, Replace: replace}
	}
	mycli := debug.Module{Path: "example.com/mycli", Version: "v9.0.0"}
	for _, c := range []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{"fejkdata at a release", &debug.BuildInfo{Main: *fejkdata("v0.1.0", nil)}, "fejkdata v0.1.0"},
		{"fejkdata built from a checkout", &debug.BuildInfo{Main: *fejkdata("(devel)", nil)}, "fejkdata devel"},
		{"a CLI requiring a release", &debug.BuildInfo{Main: mycli, Deps: []*debug.Module{fejkdata("v0.1.0", nil)}}, "mycli v9.0.0, fejkdata v0.1.0"},
		{"a CLI replacing fejkdata with a folder", &debug.BuildInfo{Main: mycli, Deps: []*debug.Module{fejkdata("v0.1.0", &debug.Module{Path: "../fejkdata"})}}, "mycli v9.0.0, fejkdata devel"},
		{"a CLI replacing fejkdata with a fork", &debug.BuildInfo{Main: mycli, Deps: []*debug.Module{fejkdata("v0.1.0", &debug.Module{Path: "example.com/fork", Version: "v0.1.1"})}}, "mycli v9.0.0, fejkdata v0.1.1"},
		{"a CLI whose modules lack fejkdata", &debug.BuildInfo{Main: mycli}, "mycli v9.0.0, fejkdata devel"},
		{"no build info", nil, "fejkdata devel"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := versionLine("mycli", c.info); got != c.want {
				t.Errorf("versionLine(mycli, %+v) = %q, want %q", c.info, got, c.want)
			}
		})
	}
}

func TestRunVersion(t *testing.T) {
	code, out, errb := runOut("", "--version")
	if code != 0 {
		t.Fatalf("--version = %d, stderr=%q", code, errb)
	}
	version, ok := strings.CutPrefix(out, "fejkdata ")
	if !ok || strings.TrimSpace(version) == "" || strings.Index(out, "\n") != len(out)-1 {
		t.Errorf("--version = %q, want the command name and a version on stdout, ending in its one newline", out)
	}
}

func TestDocsURLPinsTheReleaseItRuns(t *testing.T) {
	fejkdata := func(version string, replace *debug.Module) *debug.Module {
		return &debug.Module{Path: "github.com/larvit/fejkdata", Version: version, Replace: replace}
	}
	mycli := debug.Module{Path: "example.com/mycli", Version: "v9.0.0"}
	pinned, latest := "https://github.com/larvit/fejkdata/blob/v0.4.0/README.md", "https://github.com/larvit/fejkdata"
	for _, c := range []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{"fejkdata at a release", &debug.BuildInfo{Main: *fejkdata("v0.4.0", nil)}, pinned},
		{"fejkdata built from a checkout", &debug.BuildInfo{Main: *fejkdata("(devel)", nil)}, latest},
		{"fejkdata at a pseudo-version", &debug.BuildInfo{Main: *fejkdata("v0.4.1-0.20261108120000-abcdef123456", nil)}, latest},
		{"fejkdata built from a dirty checkout", &debug.BuildInfo{Main: *fejkdata("v0.4.0+dirty", nil)}, latest},
		{"a CLI requiring a release", &debug.BuildInfo{Main: mycli, Deps: []*debug.Module{fejkdata("v0.4.0", nil)}}, pinned},
		{"a CLI replacing fejkdata with a fork", &debug.BuildInfo{Main: mycli, Deps: []*debug.Module{fejkdata("v0.4.0", &debug.Module{Path: "example.com/fork", Version: "v0.4.1"})}}, latest},
		{"a CLI whose modules lack fejkdata", &debug.BuildInfo{Main: mycli}, latest},
		{"no build info", nil, latest},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := docsURL(c.info); got != c.want {
				t.Errorf("docsURL(%+v) = %q, want %q", c.info, got, c.want)
			}
		})
	}
}
