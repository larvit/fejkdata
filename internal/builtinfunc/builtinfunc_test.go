package builtinfunc

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
)

// TestRegistryShapes pins the builtin contract Prep relies on: every entry but adddays supplies
// prep, and args parsed at compile only behind a check.
func TestRegistryShapes(t *testing.T) {
	for name, b := range builtins {
		if b.prep == nil && name != AddDaysName {
			t.Errorf("builtin %q has no prep: Prep would call a nil func", name)
		}
		if b.arity != 0 && b.checkArgs == nil {
			t.Errorf("builtin %q parses args in prep with no check", name)
		}
	}
}

func TestCheckNamesWhatACallGetsWrong(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"nope", nil, `unknown function "nope"`},
		{"mod11", []string{"1"}, "mod11 takes 0 arguments, got 1"},
		{"hex", nil, "hex takes 1 argument, got 0"},
		{"fnr", nil, "fnr takes 1 argument, got 0"},
		{"cpr", nil, "cpr takes 1 argument, got 0"},
		{"calc", []string{"a +"}, `calc("a +"): `},
		{"calc", []string{"a", "-1"}, "calc decimals "},
		{"calc", nil, "calc takes an expression and an optional decimals count, got 0 args"},
	} {
		if err := Check(c.name, c.args); err == nil || !strings.HasPrefix(err.Error(), c.want) {
			t.Errorf("Check(%s, %q) = %v, want it to start %q", c.name, c.args, err, c.want)
		}
	}
	if err := Check("calc", []string{"a / b", "2"}); err != nil {
		t.Errorf("Check(calc) = %v, want an expression over any operand accepted", err)
	}
}

// TestArgGuardsPanic pins the guards that report a builtin arg its check should have
// rejected. No data reaches them, since Check runs before Prep, so they are called directly.
func TestArgGuardsPanic(t *testing.T) {
	for name, call := range map[string]func(){
		"atoi on an unvalidated arg":             func() { atoi("nope") },
		"atof on an unvalidated arg":             func() { atof("nope") },
		"prep on an expression that won't parse": func() { calcPrep([]string{"1 +"}) },
	} {
		mustPanic(t, name, call)
	}
}

// mustPanic fails unless call panics with invariant.Broken's phrase, which is what
// separates a reported invariant break from a silently wrong value.
func mustPanic(t *testing.T, name string, call func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil || !strings.HasPrefix(fmt.Sprint(r), invariant.Broken("")) {
			t.Errorf("%s: recovered %v, want the invariant reported as an internal error", name, r)
		}
	}()
	call()
}

func TestAsciiFoldsEveryShippedNameLetter(t *testing.T) {
	var files []string
	for _, name := range []string{"first-name.tsv", "last-name.tsv"} {
		found, err := filepath.Glob("../../data/*/" + name)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, found...)
	}
	if len(files) == 0 {
		t.Fatal("found no first-name.tsv or last-name.tsv under data/")
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range string(src) {
			if r > unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsMark(r)) && asciiFolds[r] == "" {
				t.Errorf("%s holds %q, which ascii has no entry for", file, r)
			}
		}
	}
}

func TestEveryBuiltinShowsItsCalls(t *testing.T) {
	for name, b := range builtins {
		if len(b.calls) == 0 {
			t.Errorf("%s shows no call", name)
		}
		for _, c := range b.calls {
			if !strings.HasPrefix(c, name+"(") || !strings.HasSuffix(c, ")") {
				t.Errorf("%s shows the call %q, want %s(…)", name, c, name)
			}
		}
	}
}

func TestEachRunsHoldsWhatItsBuiltinPrintsAndNoMore(t *testing.T) {
	s := drawstate.New(1, time.Time{})
	for _, c := range []struct {
		name string
		args []string
	}{
		{"digits", []string{"3"}}, {"lower", []string{"2"}}, {"upper", []string{"2"}}, {"hex", []string{"4"}},
		{"nanoid", []string{"3"}}, {"base64", []string{"1"}}, {"base64", []string{"2"}}, {"base64", []string{"3"}},
		{"base64", []string{"4"}}, {"uuid", nil}, {"ulid", nil},
	} {
		call := fmt.Sprintf("%s(%s)", c.name, strings.Join(c.args, ","))
		runs, ok := Runs(c.name, c.args)
		if !ok {
			t.Errorf("%s has no runs", call)
			continue
		}
		var want []string
		for _, r := range runs {
			for range r.N {
				want = append(want, sortedSet(r.Alphabet))
			}
		}
		seen := make([]string, len(want))
		draw := Prep(c.name, c.args)
		for range 5000 {
			v := draw(s, "", nil)
			if len(v) != len(want) {
				t.Fatalf("%s printed %q, %d characters, where its runs hold %d", call, v, len(v), len(want))
			}
			for i := range v {
				seen[i] += v[i : i+1]
			}
		}
		for i := range want {
			if got := sortedSet(seen[i]); got != want[i] {
				t.Errorf("%s at %d printed %q, where its runs hold %q", call, i, got, want[i])
			}
		}
	}
}

func sortedSet(s string) string {
	b := []byte(s)
	slices.Sort(b)
	return string(slices.Compact(b))
}

func TestOnlyTheSampleBuiltinsHaveRuns(t *testing.T) {
	var got []string
	for name := range builtins {
		if _, ok := Runs(name, []string{"1"}); ok {
			got = append(got, name)
		}
	}
	slices.Sort(got)
	if want := []string{"base64", "digits", "hex", "lower", "nanoid", "ulid", "upper", "uuid"}; !slices.Equal(got, want) {
		t.Errorf("the builtins with runs are %q, want %q", got, want)
	}
}
