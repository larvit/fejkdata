package fejkdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
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

// SQLInsert renders the record as one INSERT statement into table, identifiers in ANSI
// double quotes.
func (r *Record) SQLInsert(table string) string {
	cols := make([]string, len(r.columns))
	vals := make([]string, len(r.columns))
	for i, c := range r.columns {
		cols[i] = quoteIdent(c.Name)
		vals[i] = literal(c, sqlString, "NULL")
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", quoteIdent(table), strings.Join(cols, ", "), strings.Join(vals, ", "))
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
	segments, err := grammar.SplitPath(path)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	f.loadShippedAt(segments)
	_, n, tail, err := resolveCategory(f.root.children, segments)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	var frames frameStack
	env := renderEnv{frames: &frames}
	if t, isTable := n.(*table); isTable {
		if n, env.row, err = tableRecord(f.drawState, t, tail); err != nil {
			return nil, fmt.Errorf("fejkdata: %s: %w", path, err)
		}
	} else if len(tail) > 0 {
		return nil, fmt.Errorf("fejkdata: %s descends into %q, a field; only a category-level template is a record", path, tail[0])
	}
	record, err := recordOf(n)
	if errors.Is(err, ErrNoColumns) {
		ns := grammar.NameSegments(segments)
		return nil, fmt.Errorf(`fejkdata: %s %w; render it as a column of one: {"format":"","%s":"{/%s}"}`, path, err, ns[len(ns)-1], path)
	}
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %s %w", path, err)
	}
	return renderRecord(f.drawState, record, env), nil
}

// tableRecord walks a path's tail from a table to the table whose row is the record, and that
// row, selected or drawn.
func tableRecord(s *drawstate.State, t *table, tail []string) (node, renderedRow, error) {
	n, pins, err := descend(s, t, tail)
	if err != nil {
		return nil, renderedRow{}, err
	}
	switch n := n.(type) {
	case *table:
		return n, renderedRow{n, n.rows.DrawIn(s, &pins)}, nil
	case *tableRow:
		return n.t, renderedRow{n.t, pins.MustRow(n.t.rows)}, nil
	case *tableColumn:
		return nil, renderedRow{}, fmt.Errorf("descends into %q, a column; a record is a table's row", n.name())
	}
	return n, renderedRow{}, nil
}

// RecordTemplate is an inline record compiled, referenced and validated once,
// ready to render many times with [RecordTemplate.Fake].
type RecordTemplate struct {
	g        *Generator
	template *template
}

// Fake renders the record, one pick of each name across its columns.
func (t *RecordTemplate) Fake() *Record {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	var frames frameStack
	return renderRecord(t.g.drawState, t.template, renderEnv{frames: &frames})
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

// ErrNoColumns is the one record fence a path can answer, so its entry point names
// the record to write instead.
var ErrNoColumns = errors.New("has no fields, so no columns")

// recordOf is the record template n is; settleRecords fixed its columns at load.
func recordOf(n node) (*template, error) {
	if tb, isTable := n.(*table); isTable {
		n = tb.formatTemplate
	}
	t, ok := n.(*template)
	if !ok {
		return nil, errors.New("names a choice, not a template; a record is a template whose fields are its columns")
	}
	if len(t.fields) == 0 {
		return nil, ErrNoColumns
	}
	if !t.isRecord {
		return nil, fmt.Errorf("carries repeat %d, which composes its format into one string; a record projects columns instead — drop the repeat and render the record again for more rows", t.repeat)
	}
	if len(t.columns) != len(t.fields) {
		panic(invariant.Broken("record %q was reached before its load settled its columns", t.format))
	}
	return t, nil
}

// recordColumn is one column of a record template: its name and datatype, the field drawing it,
// and the name the template binds to that whole field, whose pick the column renders so the row
// agrees with the columns reading the name.
type recordColumn struct {
	name       string
	datatype   DataType
	field      node
	boundWhole *nameBinding
}

// settleRecords fixes the columns of every record in nodes. The pipeline runs it after checkNoCycles: a
// column's datatype walks the column, and a cycle would walk forever.
func settleRecords(nodes nodeSet) {
	_ = nodes(func(_ string, n node) error {
		if t, isTemplate := n.(*template); isTemplate {
			t.columns = recordColumns(t)
		}
		return nil
	})
}

// recordColumns is t's columns in name order where t is a record with fields, else nil.
func recordColumns(t *template) []recordColumn {
	if !t.isRecord || len(t.fields) == 0 {
		return nil
	}
	whole := map[node]*nameBinding{}
	for _, tok := range t.tokens {
		if tok.Kind != grammar.NameBind {
			continue
		}
		if b := t.nameScope.bindings[tok.Bound]; b.bindsField() && len(b.target.tail) == 0 && whole[b.target.start] == nil {
			whole[b.target.start] = b
		}
	}
	names := sortedNames(t.fields)
	columns := make([]recordColumn, len(names))
	for i, name := range names {
		field := t.fields[name]
		columns[i] = recordColumn{name: name, datatype: columnDatatype(field), field: field, boundWhole: whole[field]}
	}
	return columns
}

// renderRecord draws each column of t once, in name order, as one render, so the columns read one
// pick of each name; a table's columns read env's row.
func renderRecord(s *drawstate.State, t *template, env renderEnv) *Record {
	if mark := env.renderFrame(t); mark >= 0 {
		defer env.frames.pop(mark)
	}
	r := &Record{columns: make([]Column, len(t.columns))}
	for i, c := range t.columns {
		var column readValue
		if c.boundWhole != nil {
			column = readName(s, env, arm{kind: namedRead, named: c.boundWhole, levels: wholeLevels})
		} else {
			column = renderLeaf(s, c.field, env)
		}
		r.columns[i] = Column{Name: c.name, DataType: c.datatype, Value: column.text, Null: column.null}
	}
	return r
}

// wholeLevels is the levels of a read of a whole name, as compileArm compiles {n}.
var wholeLevels = []pickKey{""}
