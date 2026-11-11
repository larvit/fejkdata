package fejkdata

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine"
)

// Column is one rendered column of a record: its Name, its DataType and its Value, the
// rendered text, which a serializer quotes for DataTypeString and writes bare for any other
// datatype. A Null column has no Value.
type Column = engine.Column

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
	record, err := engine.RecordAt(f.drawState, &f.root, path)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return renderRecord(f.drawState, record), nil
}

// RecordTemplate is an inline record compiled, referenced and validated once,
// ready to render many times with [RecordTemplate.Fake].
type RecordTemplate struct {
	g      *Generator
	record engine.Record
}

// Fake renders the record, one pick of each name across its columns.
func (t *RecordTemplate) Fake() *Record {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return renderRecord(t.g.drawState, t.record)
}

// NewRecordTemplate compiles an inline record — a JSON object with a format and
// fields — and resolves its references against the loaded tree.
func (f *Generator) NewRecordTemplate(input string) (*RecordTemplate, error) {
	t, err := f.NewTemplate(input)
	if err != nil {
		return nil, err
	}
	record, err := t.template.Record()
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &RecordTemplate{g: f, record: record}, nil
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
// field in "let". For the first two, FakeRecord's error also names the one-column record to write
// instead.
var ErrNoColumns = engine.ErrNoColumns

// renderRecord draws r.
func renderRecord(s *drawstate.State, r engine.Record) *Record {
	out := &Record{columns: make([]Column, r.ColumnCount())}
	r.Draw(s, func(i int, c engine.Column) { out.columns[i] = c })
	return out
}
