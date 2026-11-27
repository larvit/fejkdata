//go:build js && wasm

package main

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/larvit/fejkdata"
)

// kindError is an error whose kind is known where it arises.
type kindError struct {
	kind string
	err  error
}

func (e kindError) Error() string { return e.err.Error() }

func misuse(format string, args ...any) error {
	return kindError{"misuse", fmt.Errorf(format, args...)}
}

// failure is {error, kind, exitCode} for err. retry redoes over other modules the step that gave
// err. The exit code is 2 where the CLI would call err misuse.
func (b bridge) failure(err error, retry func([]fs.FS) error) map[string]any {
	kind, note := b.kindOf(err, retry)
	exitCode := 1
	if kind == "misuse" || note != "" {
		exitCode = 2
	}
	return map[string]any{"error": err.Error() + note, "kind": kind, "exitCode": exitCode}
}

// kindOf is the kind of err, and a note to add to its message: load where the data fails to load
// or where retry passes once every shipped module is loaded, else misuse.
func (b bridge) kindOf(err error, retry func([]fs.FS) error) (string, string) {
	var known kindError
	switch {
	case errors.As(err, &known):
		return known.kind, ""
	case errors.Is(err, fejkdata.ErrLoad):
		return "load", ""
	case b.every != nil && retry(b.every) == nil:
		return "load", leftOut
	}
	return "misuse", ""
}

// compile compiles template over modules as the CLI does under format.
func compile(modules []fs.FS, format, template string) error {
	f, err := fejkdata.New(fejkdata.WithDataFS(modules...), fejkdata.WithSeed(0), fejkdata.WithToday(anyDay))
	if err != nil {
		return err
	}
	if format == "text" {
		_, err = f.NewTemplate(template)
		return err
	}
	draw, err := recordDrawer(f, template)
	if err == nil {
		_, err = draw()
	}
	return err
}
