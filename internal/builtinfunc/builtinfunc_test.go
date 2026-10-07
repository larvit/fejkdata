package builtinfunc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/invariant"
)

// TestRegistryShapes pins the builtin contract Prep relies on: every entry supplies prep,
// and args parsed at compile only behind a check.
func TestRegistryShapes(t *testing.T) {
	for name, b := range builtins {
		if b.prep == nil {
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
		{"luhn", []string{"1"}, "luhn takes 0 arguments, got 1"},
		{"hex", nil, "hex takes 1 argument, got 0"},
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
