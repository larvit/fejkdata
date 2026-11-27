//go:build js && wasm

// Command wasm is fejkdata built to WebAssembly for the npm package. It hands index.js one
// function, open, and reads only the library's public API and the CLI's Run.
package main

import (
	"errors"
	"io/fs"
	"strings"
	"syscall/js"

	"github.com/larvit/fejkdata/cli"
)

func main() {
	js.Global().Call("__fejkdataRun", js.FuncOf(open))
	select {}
}

// open(data, files) returns {render, records, paths} over the modules data names, or every
// shipped module where data is null, with files layered over them. It returns {error} where it
// refuses data or files.
func open(_ js.Value, call []js.Value) any {
	b, err := newBridge(call[0], call[1])
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"render": js.FuncOf(b.render), "records": js.FuncOf(b.records), "paths": js.FuncOf(b.paths)}
}

// bridge holds the modules one load gave.
type bridge struct {
	modules []fs.FS
	// every is each shipped module with the files, where data leaves a module out.
	every []fs.FS
}

// render(seed, today, repeat, format, template) runs the CLI, and returns {exitCode, stdout,
// stderr}, with the error's kind and a note to add to stderr where the exit code is not 0.
func (b bridge) render(_ js.Value, call []js.Value) any {
	format, template := call[3].String(), call[4].String()
	args := []string{"--seed", call[0].String(), "--today", call[1].String(), "--repeat", call[2].String(), "--format", format}
	var stdout capped
	var stderr strings.Builder
	exitCode := cli.Run(args, strings.NewReader(template), &stdout, &stderr, b.modules...)
	if exitCode == 0 {
		return map[string]any{"exitCode": 0, "stdout": stdout.text.String()}
	}
	kind, note := "limit", ""
	if !stdout.full {
		kind, note = b.renderKind(exitCode, format, template)
	}
	return map[string]any{"exitCode": exitCode, "stderr": stderr.String(), "kind": kind, "note": note}
}

// renderKind is the kind of a failed render's error: a template's or its data's where they fail
// to compile, else misuse for exit code 2 and other for any other.
func (b bridge) renderKind(exitCode int, format, template string) (string, string) {
	err := compile(b.modules, format, template)
	if err != nil {
		return b.kindOf(err, format, template)
	}
	if exitCode == 2 {
		return "misuse", ""
	}
	return "other", ""
}

// maxOutput is the shortest string limit among the JavaScript engines, V8's, in UTF-16 units. A
// text holds at least as many UTF-8 bytes as UTF-16 units, so a text within it in bytes fits.
const maxOutput = 1<<29 - 24

var errTooLong = errors.New("the output passes 512 MB, the most a JavaScript string holds; render fewer rows a call")

// capped collects what is written to it, and refuses a write that takes it past maxOutput.
type capped struct {
	text strings.Builder
	full bool
}

func (c *capped) Write(p []byte) (int, error) {
	if c.text.Len()+len(p) > maxOutput {
		c.full = true
		return 0, errTooLong
	}
	return c.text.Write(p)
}
