//go:build js && wasm

// Command wasm is fejkdata built to WebAssembly for the npm package. It hands index.js the
// functions check, render, fake, records and paths, and reads only the library's public API and
// the CLI's Run. Each function takes load's data and files first and keeps nothing after it
// returns, so the memory a load's files take is JavaScript's to free.
package main

import (
	"errors"
	"io/fs"
	"strings"
	"syscall/js"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/cli"
)

func main() {
	js.Global().Call("__fejkdataRun", map[string]any{
		"check":   js.FuncOf(opened(check)),
		"render":  js.FuncOf(opened(bridge.render)),
		"fake":    js.FuncOf(opened(bridge.fake)),
		"records": js.FuncOf(opened(bridge.records)),
		"paths":   js.FuncOf(opened(bridge.paths)),
	})
	select {}
}

// opened builds the bridge over the data and files a call starts with, and runs call over it
// with the arguments after them. It returns {error, kind, exitCode} where it refuses data or files.
func opened(call func(bridge, []js.Value) any) func(js.Value, []js.Value) any {
	return func(_ js.Value, args []js.Value) any {
		b, err := newBridge(args[0], args[1])
		if err != nil {
			return map[string]any{"error": err.Error(), "kind": "misuse", "exitCode": 2}
		}
		return call(b, args[2:])
	}
}

// check returns {} where load's files load over the modules data chose, else {error, kind,
// exitCode}.
func check(b bridge, _ []js.Value) any {
	if b.files == nil {
		return map[string]any{}
	}
	_, err := fejkdata.New(fejkdata.WithDataFS(b.modules()...), fejkdata.WithSeed(0), fejkdata.WithToday(anyDay))
	if err != nil {
		return b.failure(err, func([]fs.FS) error { return err })
	}
	return map[string]any{}
}

// bridge holds the modules a call reads: those load's data chose, then the files over them.
type bridge struct {
	chosen, files []fs.FS
	// every is each shipped module with the files, set where data names the modules.
	every []fs.FS
}

func (b bridge) modules() []fs.FS {
	return append(append([]fs.FS{}, b.chosen...), b.files...)
}

// render(seed, today, repeat, format, template) runs the CLI. It returns {exitCode, stdout} where
// the CLI exits 0, else {exitCode, stderr, kind, note}, the note to add to stderr.
func (b bridge) render(args []js.Value) any {
	r := renderArgs{args[0].String(), args[1].String(), args[2].String(), args[3].String(), args[4].String()}
	var stdout capped
	var stderr strings.Builder
	flags := []string{"--seed", r.seed, "--today", r.today, "--repeat", r.repeat, "--format", r.format}
	exitCode := cli.Run(flags, strings.NewReader(r.template), &stdout, &stderr, b.modules()...)
	if exitCode == 0 {
		return map[string]any{"exitCode": 0, "stdout": stdout.text.String()}
	}
	kind, note := "limit", ""
	if !stdout.full {
		kind, note = b.renderKind(exitCode, r)
	}
	return map[string]any{"exitCode": exitCode, "stderr": stderr.String(), "kind": kind, "note": note}
}

type renderArgs struct{ seed, today, repeat, format, template string }

var formats = map[string]bool{"text": true, "csv": true, "json": true, "ndjson": true, "pgcopy": true, "sql": true, "tsv": true}

// valid tells whether seed, today, repeat and format are values the CLI takes.
func (r renderArgs) valid() bool {
	_, settingsErr := drawSettings(r.seed, r.today)
	_, repeatErr := repeatCount(r.repeat)
	return settingsErr == nil && repeatErr == nil && formats[r.format]
}

// renderKind is a failed render's kind and note: misuse for a bad argument, kindOf's where the
// template fails to compile, else misuse for exit code 2 and other for any other.
func (b bridge) renderKind(exitCode int, r renderArgs) (string, string) {
	if !r.valid() {
		return "misuse", ""
	}
	retry := func(modules []fs.FS) error { return compile(modules, r.format, r.template) }
	if err := retry(b.modules()); err != nil {
		return b.kindOf(err, retry)
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

var errTooMany = errors.New("the rows pass 512 MB, counting 32 bytes a cell beside its text; draw fewer rows a call")

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
