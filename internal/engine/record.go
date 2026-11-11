package engine

import (
	"errors"
	"fmt"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/render"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// Column is one drawn column of a record.
type Column struct {
	Name     string
	DataType datatype.DataType
	Value    string
	Null     bool
}

// Record is a record template whose load fixed its columns, and the table row its columns
// read, if any.
type Record struct {
	template *tree.Template
	row      render.Row
}

// ColumnCount is how many columns r draws.
func (r Record) ColumnCount() int { return len(r.template.Columns()) }

// Name is the name of r's column i.
func (r Record) Name(i int) string { return r.template.Columns()[i].Name() }

// Draw draws r's columns, handing each to set in name order.
func (r Record) Draw(s *drawstate.State, set func(i int, c Column)) {
	render.Record(s, r.template, r.row, func(i int, text string, null bool) {
		c := r.template.Columns()[i]
		set(i, Column{Name: c.Name(), DataType: c.DataType(), Value: text, Null: null})
	})
}

// RecordAt walks a caller's path to the record it names, drawing the row of a table it ends on.
// Only a category-level template or a table's row is a record. A failed load matches ErrLoad.
func RecordAt(s *drawstate.State, root *tree.Folder, path string) (Record, error) {
	path, segments, err := callerPath(root, path)
	if err != nil {
		return Record{}, err
	}
	_, n, tail, err := load.ResolveCategory(root.Children(), segments)
	if err != nil {
		return Record{}, noDataNote(root, fmt.Errorf("%s: %w", path, err))
	}
	var row render.Row
	if t, isTable := n.(*tree.Table); isTable {
		if n, row, err = tableRecord(s, t, tail); err != nil {
			return Record{}, fmt.Errorf("%s: %w", path, err)
		}
	} else if len(tail) > 0 {
		return Record{}, fmt.Errorf("%s descends into %q, a field; only a category-level template is a record", path, tail[0])
	}
	record, err := recordOf(n)
	if errors.Is(err, ErrNoColumns) && err != holdsOnlyLets {
		ns := grammar.IdentifierSegments(segments)
		return Record{}, fmt.Errorf(`%s %w; render it as a column of one: {"format":"","%s":"{/%s}"}`, path, err, ns[len(ns)-1], path)
	}
	if err != nil {
		return Record{}, fmt.Errorf("%s %w", path, err)
	}
	return Record{template: record, row: row}, nil
}

// tableRecord walks a path's tail from a table to the table whose row is the record, and that
// row, selected or drawn.
func tableRecord(s *drawstate.State, t *tree.Table, tail []string) (tree.Node, render.Row, error) {
	n, pins, err := drawCallerPath(s, t, tail)
	if err != nil {
		return nil, render.Row{}, err
	}
	switch n := n.(type) {
	case *tree.Table:
		return n, render.Row{Table: n, Index: n.Rows().DrawIn(s, &pins)}, nil
	case *tree.TableRow:
		return n.Table(), render.Row{Table: n.Table(), Index: pins.MustRow(n.Table().Rows())}, nil
	case *tree.TableColumn:
		return nil, render.Row{}, fmt.Errorf("descends into %q, a column; a record is a table's row", n.Name())
	}
	return n, render.Row{}, nil
}

// ErrNoColumns matches the error for a record with no fields, naming a choice or holding every
// field in "let". For the first two, RecordAt's error also names the one-column record to write
// instead.
var ErrNoColumns = errors.New("has no fields, so no columns")

// noColumnsError keeps its own message and matches ErrNoColumns.
type noColumnsError string

func (e noColumnsError) Error() string { return string(e) }

func (noColumnsError) Is(target error) bool { return target == ErrNoColumns }

const (
	namesAChoice  = noColumnsError("names a choice, not a template; a record is a template whose fields are its columns")
	holdsOnlyLets = noColumnsError(`holds every field in "let", which writes no column; move one out of "let"`)
)

// recordOf is the record template n is; its load fixed its columns.
func recordOf(n tree.Node) (*tree.Template, error) {
	if tb, isTable := n.(*tree.Table); isTable {
		n = tb.FormatTemplate()
	}
	t, ok := n.(*tree.Template)
	if !ok {
		return nil, namesAChoice
	}
	if len(t.Fields()) == 0 {
		return nil, ErrNoColumns
	}
	if !t.IsRecord() {
		return nil, fmt.Errorf("carries repeat %d, which composes its format into one string; a record projects columns instead — drop the repeat and render the record again for more rows", t.Repeat())
	}
	if !t.Finished() {
		panic(invariant.Broken("record %q was reached before its load finished", t.Format()))
	}
	if len(t.Columns()) == 0 {
		return nil, holdsOnlyLets
	}
	return t, nil
}
