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
func (b bridge) records(args []js.Value) any {
	seed, today, repeat, template := args[0].String(), args[1].String(), args[2].String(), args[3].String()
	rows, err := drawRecords(b.modules(), seed, today, repeat, template)
	if err != nil {
		return b.failure(err, func(modules []fs.FS) error { return compile(modules, "json", template) })
	}
	return rows.js()
}

// fake(seed, today, path) returns {value}, what the library's Fake draws for path, or {error,
// kind, exitCode}.
func (b bridge) fake(args []js.Value) any {
	seed, today, path := args[0].String(), args[1].String(), args[2].String()
	value, err := fakeValue(b.modules(), seed, today, path)
	if err != nil {
		return b.failure(err, func(modules []fs.FS) error {
			_, err := fakeValue(modules, seed, today, path)
			return err
		})
	}
	return map[string]any{"value": value}
}

func fakeValue(modules []fs.FS, seedText, todayText, path string) (string, error) {
	options, err := drawSettings(seedText, todayText)
	if err != nil {
		return "", err
	}
	f, err := fejkdata.New(append(options, fejkdata.WithDataFS(modules...))...)
	if err != nil {
		return "", err
	}
	return f.Fake(path)
}

// paths() returns {paths}, what fejkdata --list prints, or {error, kind, exitCode}.
func (b bridge) paths([]js.Value) any {
	f, err := fejkdata.New(fejkdata.WithDataFS(b.modules()...), fejkdata.WithSeed(0), fejkdata.WithToday(anyDay))
	if err != nil {
		return b.failure(err, func([]fs.FS) error { return err })
	}
	var paths []any
	for _, path := range f.List() {
		paths = append(paths, path)
	}
	return map[string]any{"paths": paths}
}

// anyDay is the day a call that draws nothing gives New.
var anyDay = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func drawRecords(modules []fs.FS, seedText, todayText, repeatText, template string) (*table, error) {
	options, err := drawSettings(seedText, todayText)
	if err != nil {
		return nil, err
	}
	repeat, err := repeatCount(repeatText)
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
	rows := table{rows: repeat}
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

// drawSettings reads seed and today as the CLI reads its flags of those names.
func drawSettings(seedText, todayText string) ([]fejkdata.Option, error) {
	seed, err := strconv.ParseUint(seedText, 10, 64)
	if err != nil {
		return nil, misuse("fejkdata: seed needs an unsigned integer, got %q", seedText)
	}
	today, err := time.Parse(time.DateOnly, todayText)
	if err != nil || today.Year() < 1000 {
		return nil, misuse("fejkdata: today needs a date as YYYY-MM-DD, from year 1000 on, got %q", todayText)
	}
	return []fejkdata.Option{fejkdata.WithSeed(seed), fejkdata.WithToday(today)}, nil
}

// repeatCount reads repeat as the CLI reads --repeat.
func repeatCount(repeatText string) (int, error) {
	repeat, err := strconv.Atoi(repeatText)
	if err != nil || repeat < 0 || repeat > fejkdata.MaxRepeat {
		return 0, misuse("fejkdata: repeat needs an integer in 0..%d, got %q", fejkdata.MaxRepeat, repeatText)
	}
	return repeat, nil
}

// recordDrawer draws template's records: the path's where it is one reference from the root, as
// the CLI reads one, else the record template's.
func recordDrawer(f *fejkdata.Generator, template string) (func() (*fejkdata.Record, error), error) {
	if path, lone := recordPath(template); lone {
		return func() (*fejkdata.Record, error) { return f.FakeRecord(path) }, nil
	}
	t, err := f.NewRecordTemplate(template)
	if err != nil {
		return nil, spacedReference(template, err)
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

// spacedReference adds to err the hint the CLI's records give where whitespace around a lone
// reference made the template text with no column.
func spacedReference(template string, err error) error {
	if trimmed := strings.TrimSpace(template); trimmed != template && errors.Is(err, fejkdata.ErrNoColumns) {
		if path, lone := recordPath(trimmed); lone {
			return fmt.Errorf("%w; the whitespace around {/%s} makes it text, so remove it to write the record %s", err, path, path)
		}
	}
	return err
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
// table holds the rows records returns. size counts their text, and cellCost for each cell.
type table struct {
	names, types, values []any
	size, rows           int
}

const cellCost = 32

func (t *table) add(columns []fejkdata.Column) error {
	if t.names == nil {
		if err := t.name(columns); err != nil {
			return err
		}
	}
	for _, c := range columns {
		t.size += cellCost + len(c.Value)
		if t.size > maxOutput {
			return kindError{"limit", errTooMany}
		}
		if c.Null {
			t.values = append(t.values, nil)
		} else {
			t.values = append(t.values, c.Value)
		}
	}
	return nil
}

// name takes the columns' names and types, and refuses rows whose cells alone pass maxOutput.
func (t *table) name(columns []fejkdata.Column) error {
	for _, c := range columns {
		t.names = append(t.names, c.Name)
		t.types = append(t.types, c.DataType.String())
		t.size += 2*cellCost + len(c.Name) + len(c.DataType.String())
	}
	if t.size+t.rows*len(columns)*cellCost > maxOutput {
		return kindError{"limit", errTooMany}
	}
	return nil
}

func (t *table) js() map[string]any {
	return map[string]any{"names": t.names, "types": t.types, "values": t.values}
}
