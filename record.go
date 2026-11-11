package fejkdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/render"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// Column is one rendered column of a record. Value is the rendered text, which a
// serializer quotes for DataTypeString and writes bare for any other datatype; a Null
// column has no Value.
type Column struct {
	Name     string
	DataType DataType
	Value    string
	Null     bool
}

// Record is one record rendered from a template: every direct field is a column,
// listed in name order. Each {…} draws afresh, while a pick bound by {x as n} is
// drawn once for the whole record.
type Record struct {
	columns []Column
}

// Columns returns a copy of the record's columns, in name order.
func (r *Record) Columns() []Column {
	return append([]Column(nil), r.columns...)
}

// JSON renders the record as one JSON object.
func (r *Record) JSON() string {
	var b strings.Builder
	b.WriteByte('{')
	for i, c := range r.columns {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(jsonString(c.Name))
		b.WriteByte(':')
		b.WriteString(literal(c, jsonString, "null"))
	}
	b.WriteByte('}')
	return b.String()
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// CSVHeader renders the column names as one CSV header line.
func (r *Record) CSVHeader() string {
	fields := make([]string, len(r.columns))
	for i, c := range r.columns {
		fields[i] = csvField(c.Name)
	}
	return strings.Join(fields, ",")
}

// CSVLine renders the column values as one CSV row: a null column an empty field and an
// empty string "", the convention PostgreSQL's COPY reads a null by. A record of one null
// column is a blank line, which COPY reads as null and most CSV readers skip.
func (r *Record) CSVLine() string {
	fields := make([]string, len(r.columns))
	for i, c := range r.columns {
		fields[i] = literal(c, csvField, "")
	}
	return strings.Join(fields, ",")
}

// csvField quotes a field where encoding/csv would, and an empty one too.
func csvField(s string) string {
	first, _ := utf8.DecodeRuneInString(s)
	switch {
	case s == "":
		return `""`
	case s == `\.` || strings.ContainsAny(s, "\",\r\n") || unicode.IsSpace(first):
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// TSVHeader renders the column names as one tab-separated line, the header of a table.
func (r *Record) TSVHeader() string {
	names := make([]string, len(r.columns))
	for i, c := range r.columns {
		names[i] = tsvCell(c.Name)
	}
	return strings.Join(names, "\t")
}

// TSVLine renders the column values as one table row. A null column and an empty string are an empty
// cell, a brace is doubled, as a table cell reads braces as tokens, and a tab, CR or LF is a space.
func (r *Record) TSVLine() string {
	cells := make([]string, len(r.columns))
	for i, c := range r.columns {
		cells[i] = literal(c, tsvCell, "")
	}
	return strings.Join(cells, "\t")
}

// tsvCell spells a tab or line break as a space, as a table has no escape for one and a render must not fail mid-run.
func tsvCell(s string) string { return tsvCellEscapes.Replace(s) }

var tsvCellEscapes = strings.NewReplacer("{", "{{", "}", "}}", "\t", " ", "\r", " ", "\n", " ")

// SQLInsert renders the record as one INSERT statement into table, identifiers in ANSI
// double quotes.
func (r *Record) SQLInsert(table string) string {
	return r.SQLInsertHeader(table) + " " + r.SQLValues() + ";"
}

// SQLInsertHeader renders the start of an INSERT into table, up to and including VALUES. Follow
// it with records' SQLValues, a comma between each, and a semicolon.
func (r *Record) SQLInsertHeader(table string) string {
	cols := make([]string, len(r.columns))
	for i, c := range r.columns {
		cols[i] = quoteIdent(c.Name)
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES", quoteIdent(table), strings.Join(cols, ", "))
}

// SQLValues renders the record as one row of a VALUES list.
func (r *Record) SQLValues() string {
	vals := make([]string, len(r.columns))
	for i, c := range r.columns {
		vals[i] = literal(c, sqlString, "NULL")
	}
	return "(" + strings.Join(vals, ", ") + ")"
}

func sqlString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// literal writes a string column quoted, a proven typed one bare, and a null as nullText.
func literal(c Column, quote func(string) string, nullText string) string {
	switch {
	case c.Null:
		return nullText
	case c.DataType == DataTypeString:
		return quote(c.Value)
	}
	return c.Value
}

// FakeRecord renders a path as one record: the template it names, with each direct
// field drawn as a column. Only a category-level template is a record — a path
// that descends into a field, or that names a folder or a choice, is an error.
func (f *Generator) FakeRecord(path string) (*Record, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path, segments, err := f.loadCallerPath(path)
	if err != nil {
		return nil, err
	}
	_, n, tail, err := load.ResolveCategory(f.root.Children(), segments)
	if err != nil {
		return nil, f.noDataNote(fmt.Errorf("fejkdata: %s: %w", path, err))
	}
	var row render.Row
	if t, isTable := n.(*tree.Table); isTable {
		if n, row, err = tableRecord(f.drawState, t, tail); err != nil {
			return nil, fmt.Errorf("fejkdata: %s: %w", path, err)
		}
	} else if len(tail) > 0 {
		return nil, fmt.Errorf("fejkdata: %s descends into %q, a field; only a category-level template is a record", path, tail[0])
	}
	record, err := recordOf(n)
	if errors.Is(err, ErrNoColumns) && err != holdsOnlyLets {
		ns := grammar.IdentifierSegments(segments)
		return nil, fmt.Errorf(`fejkdata: %s %w; render it as a column of one: {"format":"","%s":"{/%s}"}`, path, err, ns[len(ns)-1], path)
	}
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %s %w", path, err)
	}
	return renderRecord(f.drawState, record, row), nil
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

// RecordTemplate is an inline record compiled, referenced and validated once,
// ready to render many times with [RecordTemplate.Fake].
type RecordTemplate struct {
	g        *Generator
	template *tree.Template
}

// Fake renders the record, one pick of each name across its columns.
func (t *RecordTemplate) Fake() *Record {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return renderRecord(t.g.drawState, t.template, render.Row{})
}

// NewRecordTemplate compiles an inline record — a JSON object with a format and
// fields — and resolves its references against the loaded tree.
func (f *Generator) NewRecordTemplate(input string) (*RecordTemplate, error) {
	t, err := f.NewTemplate(input)
	if err != nil {
		return nil, err
	}
	record, err := recordOf(t.n)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: an inline record %w", err)
	}
	return &RecordTemplate{g: f, template: record}, nil
}

// FakeRecordTemplate compiles and renders an inline record in one call.
func (f *Generator) FakeRecordTemplate(input string) (*Record, error) {
	t, err := f.NewRecordTemplate(input)
	if err != nil {
		return nil, err
	}
	return t.Fake(), nil
}

// ErrNoColumns matches the error for a record with no fields, naming a choice or holding every
// field in "let"; FakeRecord's names the one-column record to write instead, for all but the last.
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

// settleRecords fixes the columns of every record in nodes. The pipeline runs it after load.CheckNoCycles: a
// column's datatype walks the column, and a cycle would walk forever.
func settleRecords(nodes nodeSet, l *tree.LoadFinds) {
	_ = nodes(func(_ string, n tree.Node) error {
		if t, isTemplate := n.(*tree.Template); isTemplate {
			l.SetColumns(t, recordColumns(t, l))
		}
		return nil
	})
}

// recordColumns is t's columns in name order where t is a record with fields, else nil.
func recordColumns(t *tree.Template, l *tree.LoadFinds) []tree.RecordColumn {
	if !t.IsRecord() || len(t.Fields()) == 0 {
		return nil
	}
	whole := map[tree.Node]*tree.NameBinding{}
	for _, tok := range t.Tokens() {
		if tok.Kind != grammar.NameBind {
			continue
		}
		b := l.Of(t).NameScope().Bindings()[tok.Bound]
		if target := l.Binding(b).Target(); b.BindsField() && len(target.Tail()) == 0 && whole[target.Start()] == nil {
			whole[target.Start()] = b
		}
	}
	var columns []tree.RecordColumn
	for _, name := range tree.SortedNames(t.Fields()) {
		if !t.Lets()[name] {
			field := t.Fields()[name]
			columns = append(columns, tree.NewRecordColumn(name, columnDatatype(field, l), field, whole[field]))
		}
	}
	return columns
}

// renderRecord draws t as a record; a table's columns read row.
func renderRecord(s *drawstate.State, t *tree.Template, row render.Row) *Record {
	r := &Record{columns: make([]Column, len(t.Columns()))}
	render.Record(s, t, row, func(i int, text string, null bool) {
		c := t.Columns()[i]
		r.columns[i] = Column{Name: c.Name(), DataType: c.DataType(), Value: text, Null: null}
	})
	return r
}
