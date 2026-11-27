//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/larvit/fejkdata"
)

// records(seed, today, repeat, template) returns {names, types, values}: each column's name and
// datatype, and every row's values in turn, a null as null. It returns {error, kind, exitCode}
// where it fails.
func (b bridge) records(_ js.Value, call []js.Value) any {
	template := call[3].String()
	rows, err := drawRecords(b.modules, call[0].String(), call[1].String(), call[2].String(), template)
	if err != nil {
		return b.failure(err, "json", template)
	}
	return rows.js()
}

// paths() returns {paths}, what fejkdata --list prints, or {error, kind, exitCode}.
func (b bridge) paths(js.Value, []js.Value) any {
	f, err := fejkdata.New(fejkdata.WithDataFS(b.modules...), fejkdata.WithSeed(0), fejkdata.WithToday(anyDay))
	if err != nil {
		return b.failure(err, "text", "")
	}
	var paths []any
	for _, path := range f.List() {
		paths = append(paths, path)
	}
	return map[string]any{"paths": paths}
}

// anyDay is the day a call that draws nothing gives New.
var anyDay = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// kindError is an error whose kind is known where it arises.
type kindError struct {
	kind string
	err  error
}

func (e kindError) Error() string { return e.err.Error() }

func misuse(format string, args ...any) error {
	return kindError{"misuse", fmt.Errorf(format, args...)}
}

// failure is {error, kind, exitCode} for err, which compiling template under format gave or
// which arose before it compiled. Its exit code is 2 where the CLI would call it misuse.
func (b bridge) failure(err error, format, template string) map[string]any {
	kind, note := b.kindOf(err, format, template)
	exitCode := 1
	if kind == "misuse" || note != "" {
		exitCode = 2
	}
	return map[string]any{"error": err.Error() + note, "kind": kind, "exitCode": exitCode}
}

// kindOf is the kind of err, and a note to add to its message: load where the data fails to load
// or where the template compiles once every shipped module is loaded, else misuse.
func (b bridge) kindOf(err error, format, template string) (string, string) {
	var known kindError
	switch {
	case errors.As(err, &known):
		return known.kind, ""
	case errors.Is(err, fejkdata.ErrLoad):
		return "load", ""
	case b.every != nil && compile(b.every, format, template) == nil:
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

func drawRecords(modules []fs.FS, seedText, todayText, repeatText, template string) (*table, error) {
	options, repeat, err := drawOptions(seedText, todayText, repeatText)
	if err != nil {
		return nil, err
	}
	f, err := fejkdata.New(append(options, fejkdata.WithDataFS(modules...))...)
	if err != nil {
		return nil, err
	}
	draw, err := recordDrawer(f, template)
	if err != nil {
		return nil, err
	}
	if repeat == 0 {
		_, err := draw()
		return &table{}, err
	}
	return drawRows(draw, repeat)
}

func drawRows(draw func() (*fejkdata.Record, error), repeat int) (*table, error) {
	var rows table
	for range repeat {
		r, err := draw()
		if err != nil {
			return nil, err
		}
		if err := rows.add(r.Columns()); err != nil {
			return nil, err
		}
	}
	return &rows, nil
}

// drawOptions reads records' seed, today and repeat as the CLI reads its flags of those names.
func drawOptions(seedText, todayText, repeatText string) ([]fejkdata.Option, int, error) {
	seed, err := strconv.ParseUint(seedText, 10, 64)
	if err != nil {
		return nil, 0, misuse("fejkdata: seed needs an unsigned integer, got %q", seedText)
	}
	today, err := time.Parse(time.DateOnly, todayText)
	if err != nil || today.Year() < 1000 {
		return nil, 0, misuse("fejkdata: today needs a date as YYYY-MM-DD, from year 1000 on, got %q", todayText)
	}
	repeat, err := strconv.Atoi(repeatText)
	if err != nil || repeat < 0 || repeat > fejkdata.MaxRepeat {
		return nil, 0, misuse("fejkdata: repeat needs an integer in 0..%d, got %q", fejkdata.MaxRepeat, repeatText)
	}
	return []fejkdata.Option{fejkdata.WithSeed(seed), fejkdata.WithToday(today)}, repeat, nil
}

// recordDrawer draws template's records: the path's where it is one reference from the root, as
// the CLI reads one, else the record template's.
func recordDrawer(f *fejkdata.Generator, template string) (func() (*fejkdata.Record, error), error) {
	if path, lone := recordPath(template); lone {
		return func() (*fejkdata.Record, error) { return f.FakeRecord(path) }, nil
	}
	t, err := f.NewRecordTemplate(template)
	if err != nil {
		return nil, err
	}
	return func() (*fejkdata.Record, error) { return t.Fake(), nil }, nil
}

// recordPath is the path template reads where it is {/path} and nothing else, bare or as a JSON
// string, once JSON's surrounding whitespace, or one final newline, is cut.
func recordPath(template string) (string, bool) {
	text := template
	if trimmed := strings.Trim(template, " \t\r\n"); json.Valid([]byte(trimmed)) {
		text = trimmed
	}
	if cut, found := strings.CutSuffix(text, "\n"); found {
		text = strings.TrimSuffix(cut, "\r")
	}
	path, opens := strings.CutPrefix(unquoted(text), "{/")
	path, closes := strings.CutSuffix(path, "}")
	return path, opens && closes && fejkdata.IsLonePath(path)
}

// unquoted is the string text holds where it is a JSON string, else text.
func unquoted(text string) string {
	var s string
	if strings.HasPrefix(text, `"`) && json.Unmarshal([]byte(text), &s) == nil {
		return s
	}
	return text
}

// table is the rows of records, held as their column names and datatypes and every value in turn.
type table struct {
	names, types, values []any
	size                 int
}

func (t *table) add(columns []fejkdata.Column) error {
	if t.names == nil {
		t.name(columns)
	}
	for _, c := range columns {
		t.size += len(c.Value)
		if t.size > maxOutput {
			return kindError{"limit", errTooLong}
		}
		if c.Null {
			t.values = append(t.values, nil)
		} else {
			t.values = append(t.values, c.Value)
		}
	}
	return nil
}

func (t *table) name(columns []fejkdata.Column) {
	for _, c := range columns {
		t.names = append(t.names, c.Name)
		t.types = append(t.types, c.DataType.String())
	}
}

func (t *table) js() map[string]any {
	return map[string]any{"names": t.names, "types": t.types, "values": t.values}
}
