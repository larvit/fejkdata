package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTakesNoOperand(t *testing.T) {
	file := filepath.Join(t.TempDir(), "users.tmpl")
	if err := os.WriteFile(file, []byte("{/sv_SE.person}"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"sv_SE.person"}, "echo '{/sv_SE.person}' | fejkdata"},
		{[]string{"/en_US.address"}, "echo '{/en_US.address}' | fejkdata"},
		{[]string{"{/en_US.person}"}, "echo '{/en_US.person}' | fejkdata"},
		{[]string{"name: {/sv_SE.person.last}"}, "echo 'name: {/sv_SE.person.last}' | fejkdata"},
		{[]string{"geo.US.locality[O'Fallon].name"}, `echo '{/geo.US.locality[O'\''Fallon].name}' | fejkdata`},
		{[]string{"en_US.address"}, "echo '{/en_US.address}' | fejkdata"},
		{[]string{`C:\temp {/sv_SE.person}`}, `printf '%s' 'C:\temp {/sv_SE.person}' | fejkdata`},
		{[]string{"-"}, "drop it"},
		{[]string{file}, "fejkdata < " + file},
	} {
		t.Run(argsName(c.args), func(t *testing.T) {
			code, out, errb := runOut("", c.args...)
			if code != 2 || out != "" || !strings.Contains(errb, "unexpected argument") || !strings.Contains(errb, c.want) {
				t.Errorf("run(%v) = %d, %q, %q; want misuse naming %q", c.args, code, out, errb, c.want)
			}
		})
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("bad file descriptor") }

func TestRunUnreadableStdinIsMisuse(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run("fejkdata", shipped, nil, system{1, testDay}, failingReader{}, &out, &errb); code != 2 || !strings.Contains(errb.String(), "cannot be read") {
		t.Errorf("run with an unreadable stdin = %d, %q; want misuse naming it", code, errb.String())
	}
}

func TestRunTerminalStdinIsMisuse(t *testing.T) {
	code, out, errb := runTerminal("--seed", "1")
	if code != 2 || out != "" || !strings.Contains(errb, "nothing is piped in: echo 'id: {int(1,999)}' | fejkdata; fejkdata --list prints the paths a {/path} reference reads\n") {
		t.Errorf("run with nothing piped = %d, %q, %q; want misuse showing a template piped to fejkdata", code, out, errb)
	}
}

func TestRunOperandMisuseBeforeLoad(t *testing.T) {
	code, _, errb := runOut("", "--no-shipped-data", "x")
	if code != 2 || !strings.Contains(errb, "unexpected argument") || strings.Contains(errb, "--data-path") {
		t.Fatalf("an operand with no data = %d, %q; want the operand refused before any load", code, errb)
	}
}

func TestRunInlineTemplate(t *testing.T) {
	code, out, errb := runOut("name: {/sv_SE.person.last}", "--seed", "1")
	if code != 0 || !strings.HasPrefix(out, "name: ") || strings.Contains(out, "{") {
		t.Fatalf("inline format string = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut(`{"format":"name: {x}","x":["bosse","lina"]}`, "--seed", "1")
	if code != 0 || (out != "name: bosse" && out != "name: lina") {
		t.Fatalf("inline JSON template = %d, %q, want one name, stderr %q", code, out, errb)
	}
	code, out, errb = runOut(`"name: {/sv_SE.person.last}"`, "--seed", "1")
	if code != 0 || !strings.HasPrefix(out, "name: ") || strings.Contains(out, "{") {
		t.Fatalf("inline JSON string = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut("{digits(1)}\n", "--seed", "1", "-n", "2")
	if code != 0 || len(strings.Split(strings.TrimRight(out, "\n"), "\n")) != 2 {
		t.Fatalf("inline template with --repeat = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut("{/sv_SE.person} hihi", "--seed", "1")
	if code != 0 || !strings.HasSuffix(out, " hihi") || strings.Contains(out, "{") {
		t.Fatalf("a reference and text = %d, %q, stderr %q", code, out, errb)
	}
}

func TestRunFormatTakesOnlyALoneReferenceAsARecord(t *testing.T) {
	for _, stdin := range []string{"{//sv_SE.person}", "{/sv_SE.person|/misc.uuid}", "{/sv_SE.person as p}", "{/}", "{/.sv_SE.person}"} {
		t.Run(stdin, func(t *testing.T) {
			if code, out, errb := runOut(stdin, "--format", "json"); code != 2 || out != "" {
				t.Errorf("run(%q, --format json) = %d, %q, %q; want misuse", stdin, code, out, errb)
			}
		})
	}
	if code, out, errb := runOut(`{"format":"x"}`); code != 0 || out != "x" {
		t.Errorf(`run({"format":"x"}) = %d, %q, %q; want x`, code, out, errb)
	}
}

func TestRunTemplateMisuse(t *testing.T) {
	for stdin, want := range map[string]string{
		"{bad":                  "unterminated",
		"name: {/no.such.path}": "no entry",
		"a } b":                 "}}",
		"{//sv_SE.person}":      "write {/sv_SE.person}",
	} {
		t.Run(stdin, func(t *testing.T) {
			code, out, errb := runOut(stdin, "--seed", "1")
			if code != 2 || out != "" || !strings.Contains(errb, "try 'fejkdata --help'") || !strings.Contains(errb, want) {
				t.Errorf("run(%q) = %d, %q, %q; want misuse naming %q and --help", stdin, code, out, errb, want)
			}
			if strings.Contains(errb, "fejkdata: fejkdata:") {
				t.Errorf("run(%q) doubled the program prefix: %q", stdin, errb)
			}
		})
	}
}

func TestRunReadsABareNumberAsAFormatString(t *testing.T) {
	if code, out, errb := runOut("42\n", "-n", "3"); code != 0 || out != "42\n42\n42\n" {
		t.Errorf("echo 42 | fejkdata -n 3 = %d, %q, %q, want 42 on three lines", code, out, errb)
	}
}

func TestRunRendersAnyJSONValue(t *testing.T) {
	for stdin, want := range map[string]string{"42": "42", "42\n": "42\n", "true\n": "true\n", " 7 ": " 7 ", "null\n": "", " null ": ""} {
		t.Run(fmt.Sprintf("%q", stdin), func(t *testing.T) {
			if code, out, errb := runOut(stdin); code != 0 || out != want {
				t.Errorf("run(%q) = %d, %q, %q, want %q", stdin, code, out, errb, want)
			}
		})
	}
}
