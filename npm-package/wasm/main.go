//go:build js && wasm

// Command wasm is the fejkdata CLI built to WebAssembly for the npm package: it hands
// index.js one function, which runs the CLI once over every shipped module.
package main

import (
	"errors"
	"strings"
	"syscall/js"

	"github.com/larvit/fejkdata/cli"
	"github.com/larvit/fejkdata/data"
)

func main() {
	js.Global().Call("__fejkdataRun", js.FuncOf(run))
	select {}
}

// run(args, stdin) runs the CLI with the array of arguments args and the template stdin, and
// returns {exitCode, stdout, stderr}.
func run(_ js.Value, call []js.Value) any {
	args := make([]string, call[0].Length())
	for i := range args {
		args[i] = call[0].Index(i).String()
	}
	var stdout capped
	var stderr strings.Builder
	exitCode := cli.Run(args, strings.NewReader(call[1].String()), &stdout, &stderr, data.Modules()...)
	return map[string]any{"exitCode": exitCode, "stdout": stdout.text.String(), "stderr": stderr.String()}
}

// maxOutput is the shortest string limit among the JavaScript engines, V8's, in UTF-16 units. A
// text holds at least as many UTF-8 bytes as UTF-16 units, so a text within it in bytes fits.
const maxOutput = 1<<29 - 24

var errTooLong = errors.New("the output passes 512 MB, the most a JavaScript string holds; render fewer rows a call")

// capped collects what is written to it, and refuses a write that takes it past maxOutput.
type capped struct{ text strings.Builder }

func (c *capped) Write(p []byte) (int, error) {
	if c.text.Len()+len(p) > maxOutput {
		return 0, errTooLong
	}
	return c.text.Write(p)
}
