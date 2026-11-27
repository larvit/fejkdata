//go:build js && wasm

// Command wasm is the fejkdata CLI built to WebAssembly for the npm package: it hands
// index.js one function, which runs the CLI once over every shipped module.
package main

import (
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
	var stdout, stderr strings.Builder
	exitCode := cli.Run(args, strings.NewReader(call[1].String()), &stdout, &stderr, data.Modules()...)
	return map[string]any{"exitCode": exitCode, "stdout": stdout.String(), "stderr": stderr.String()}
}
