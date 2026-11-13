package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/larvit/fejkdata"
)

// operandError refuses an argument that is no flag, naming the spelling that reads what
// it most likely means: dropping it for -, a file for a file's name, else a template
// piped in.
func operandError(arg, name string) error {
	unexpected := "unexpected argument '" + arg + "'"
	if arg == "-" {
		return fmt.Errorf("%s; stdin is read already, so drop it", unexpected)
	}
	if fi, err := os.Stat(arg); err == nil && fi.Mode().IsRegular() {
		return fmt.Errorf("%s; read a file from stdin: %s < %s", unexpected, name, shellQuoted(arg))
	}
	ref := "{/" + strings.TrimPrefix(arg, "/") + "}"
	_, argLone := loneReference(arg)
	inline, err := fejkdata.IsTemplate(arg)
	_, refLone := loneReference(ref)
	example := pipeHint(name)
	switch {
	case argLone || inline:
		example = piped(arg, name)
	case err == nil && refLone:
		example = piped(ref, name)
	}
	return fmt.Errorf("%s; the template is read from stdin: %s", unexpected, example)
}

// inputError marks a failure that is the input's own fault — nothing piped in, a template
// that does not compile, or a stdin that cannot be read. run reports it as misuse (exit 2, with a
// pointer to --help), unlike a lone reference to nothing under --format, which is a runtime
// error (exit 1).
type inputError struct{ error }

func (e inputError) Unwrap() error { return e.error }

// jsonSpace is the whitespace JSON allows around a value.
const jsonSpace = " \t\r\n"

// input is what stdin holds: the template, and the path a lone reference names as a
// record.
type input struct {
	template string
	record   string
}

// parseInput keeps stdin whole as the template, and names a record where stdin is a lone
// reference once JSON's surrounding whitespace, or one final newline, is cut.
func parseInput(raw string) input {
	trimmed := strings.Trim(raw, jsonSpace)
	text := raw
	if json.Valid([]byte(trimmed)) {
		text = trimmed
	}
	text, found := strings.CutSuffix(text, "\n")
	if found {
		text = strings.TrimSuffix(text, "\r")
	}
	record, _ := loneReference(text)
	return input{template: raw, record: record}
}

// loneReference returns the path that text reads when it is one reference from the root
// and nothing else, bare or as a JSON string.
func loneReference(text string) (string, bool) {
	var unquoted string
	if strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`) && json.Unmarshal([]byte(text), &unquoted) == nil {
		text = unquoted
	}
	path, opens := strings.CutPrefix(text, "{/")
	path, closes := strings.CutSuffix(path, "}")
	if !opens || !closes || !fejkdata.IsLonePath(path) {
		return "", false
	}
	return path, true
}

// spacedReference adds to a record refusal that whitespace around a lone reference made
// the template text.
func spacedReference(template string, err error) error {
	if trimmed := strings.TrimSpace(template); trimmed != template && errors.Is(err, fejkdata.ErrNoColumns) {
		if path, lone := loneReference(trimmed); lone {
			return fmt.Errorf("%w; the whitespace around {/%s} makes it text, so remove it to write the record %s", err, path, path)
		}
	}
	return err
}

// piped is the command piping template into the command called name; some shells' echo
// reads a backslash as an escape, so a template holding one goes through printf.
func piped(template, name string) string {
	if strings.ContainsRune(template, '\\') {
		return "printf '%s' " + shellQuoted(template) + " | " + name
	}
	return "echo " + shellQuoted(template) + " | " + name
}

// shellQuoted is s as a shell reads it back: bare where it holds nothing a shell expands.
func shellQuoted(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\r\n'\"$`\\*?[]{}()<>|&;#~!") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func readStdin(stdin io.Reader, name string) (input, error) {
	if stdin == nil {
		return input{}, inputError{fmt.Errorf("the template comes from stdin, and nothing is piped in: %s; %s --list prints the paths a {/path} reference reads", pipeHint(name), name)}
	}
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return input{}, inputError{fmt.Errorf("stdin cannot be read: %w", err)}
	}
	return parseInput(string(raw)), nil
}

// templateError is misuse unless the data the template reads failed to load.
func templateError(err error) error {
	if errors.Is(err, fejkdata.ErrLoad) {
		return err
	}
	return inputError{err}
}
