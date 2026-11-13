package cli

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/larvit/fejkdata"
)

func fail(stderr io.Writer, name string, err error) int {
	var te inputError
	if errors.As(err, &te) {
		return misuse(stderr, name, te.error)
	}
	report(stderr, name, err)
	if errors.Is(err, fejkdata.ErrNoColumns) {
		fmt.Fprintln(stderr, "wrap that JSON in single quotes, which keep a shell from expanding its braces")
	}
	return 1
}

func misuse(stderr io.Writer, name string, err error) int {
	report(stderr, name, err)
	fmt.Fprintf(stderr, "try '%s --help'\n", name)
	return 2
}

// report writes err under the command name, in place of the library's own on each line.
func report(stderr io.Writer, name string, err error) {
	msg := strings.ReplaceAll(strings.TrimPrefix(err.Error(), "fejkdata: "), "\nfejkdata: ", "\n"+name+": ")
	fmt.Fprintf(stderr, "%s: %s\n", name, msg)
}

// fejkdataModule is the path of the module holding the library; the root package's path is
// the module's.
var fejkdataModule = reflect.TypeOf((*fejkdata.Generator)(nil)).Elem().PkgPath()

// versionLine names the fejkdata version the binary is built with. Where another module's
// main package carries the CLI, the main module's version comes first, under the command name.
func versionLine(name string, info *debug.BuildInfo) string {
	if info == nil {
		return "fejkdata devel"
	}
	if info.Main.Path == fejkdataModule {
		return "fejkdata " + moduleVersion(info.Main)
	}
	for _, m := range info.Deps {
		if m.Path == fejkdataModule {
			return name + " " + moduleVersion(info.Main) + ", fejkdata " + moduleVersion(*m)
		}
	}
	return name + " " + moduleVersion(info.Main) + ", fejkdata devel"
}

var release = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

// docsURL is the README of the fejkdata release the binary is built with, or the latest
// README where it is built from anything other than a release.
func docsURL(info *debug.BuildInfo) string {
	if info != nil {
		mods := append([]*debug.Module{&info.Main}, info.Deps...)
		for _, m := range mods {
			if m.Path == fejkdataModule && m.Replace == nil && release.MatchString(m.Version) {
				return "https://github.com/larvit/fejkdata/blob/" + m.Version + "/README.md"
			}
		}
	}
	return "https://github.com/larvit/fejkdata"
}

// moduleVersion is the version m is built at: the replacement's where a replace directive
// names one, and "devel" where it carries none, as a local directory does.
func moduleVersion(m debug.Module) string {
	if m.Replace != nil {
		m = *m.Replace
	}
	if m.Version == "" || m.Version == "(devel)" {
		return "devel"
	}
	return m.Version
}
