package fejkdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
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
// listed in name order. Each column is its own expansion, so a sibling field is
// local to it, while a reference that reads a path is drawn once for the whole
// record, per group.
type Record struct {
	columns []Column
}

// Columns returns the record's columns in name order.
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
	segments, err := splitPath(path)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	_, n, tail, err := resolveCategory(f.categories, segments)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	set := eagerHoldSet()
	sc := renderScope{set: &set}
	if t, isTable := n.(*table); isTable {
		if n, err = tableRecord(f.rand, t, tail, sc); err != nil {
			return nil, fmt.Errorf("fejkdata: %s: %w", path, err)
		}
	} else if len(tail) > 0 {
		return nil, fmt.Errorf("fejkdata: %s descends into %q, a field; only a category-level template is a record", path, tail[0])
	}
	shape := f.recordShapeOf(n)
	if errors.Is(shape.err, ErrNoColumns) {
		ns := names(segments)
		return nil, fmt.Errorf(`fejkdata: %s %w; render it as a column of one: {"format":"","%s":"{/%s}"}`, path, shape.err, ns[len(ns)-1], path)
	}
	if shape.err != nil {
		return nil, fmt.Errorf("fejkdata: %s %w", path, shape.err)
	}
	return renderRecord(f.rand, shape.t, shape.columns, sc), nil
}

// tableRecord walks a path's tail from a table to the table whose row is the record,
// pinning the rows it selects or draws.
func tableRecord(s *session, t *table, tail []string, sc renderScope) (node, error) {
	n, err := descend(s, t, tail, sc)
	if err != nil {
		return nil, err
	}
	switch n := n.(type) {
	case *table:
		n.drawIn(s, &sc.hold().pins)
		return n, nil
	case *row:
		return n.t, nil
	case *column:
		return nil, fmt.Errorf("descends into %q, a column; a record is a table's row", n.t.columns[n.i])
	}
	return n, nil
}

// recordShape is what recordOf settled about a node: the template to project, its
// columns, or why it is not a record.
type recordShape struct {
	t       *template
	columns []Column
	err     error
}

// recordShapeOf fences a node once and remembers the answer. Callers hold the
// generator's lock.
func (f *Generator) recordShapeOf(n node) recordShape {
	if shape, done := f.records[n]; done {
		return shape
	}
	t, columns, err := recordOf(n)
	shape := recordShape{t: t, columns: columns, err: err}
	if f.records == nil {
		f.records = map[node]recordShape{}
	}
	f.records[n] = shape
	return shape
}

// RecordTemplate is an inline record compiled, referenced and validated once,
// ready to render many times with [RecordTemplate.Fake].
type RecordTemplate struct {
	g       *Generator
	t       *template
	columns []Column
}

// Fake renders the record with one draw.
func (t *RecordTemplate) Fake() *Record {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	set := eagerHoldSet()
	return renderRecord(t.g.rand, t.t, t.columns, renderScope{set: &set})
}

// NewRecordTemplate compiles an inline record — a JSON object with a format and
// fields — and binds its references against the loaded tree.
func (f *Generator) NewRecordTemplate(input string) (*RecordTemplate, error) {
	t, err := f.NewTemplate(input)
	if err != nil {
		return nil, err
	}
	tm, columns, err := recordOf(t.n)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: an inline record %w", err)
	}
	return &RecordTemplate{g: f, t: tm, columns: columns}, nil
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

// recordOf is the fence both record entry points pass. The columns come back with
// the template, fixed for every draw the caller goes on to make.
func recordOf(n node) (*template, []Column, error) {
	if tb, isTable := n.(*table); isTable {
		n = tb.format
	}
	t, ok := n.(*template)
	if !ok {
		return nil, nil, errors.New("names a choice, not a template; a record is a template whose fields are its columns")
	}
	names := recordColumns(t)
	if len(names) == 0 {
		return nil, nil, ErrNoColumns
	}
	if !t.record {
		return nil, nil, fmt.Errorf("carries repeat %d, which composes its format into one string; a record projects columns instead — drop the repeat and render the record again for more rows", t.repeat)
	}
	columns := make([]Column, len(names))
	for i, name := range names {
		datatype, _ := columnDatatype(t.fields[name]) // checkColumns refused items that disagree wherever DataType is read
		columns[i] = Column{Name: name, DataType: datatype}
	}
	return t, columns, nil
}

// renderRecord draws each column once, in the name order recordOf fixed, as one render
// over sc's hold; a table's columns read the row pinned there.
func renderRecord(s *session, t *template, columns []Column, sc renderScope) *Record {
	sc = sc.in(t)
	if t.table != nil {
		sc.t, sc.row = t.table, sc.hold().pins.mustRow(t.table)
	}
	r := &Record{columns: append([]Column(nil), columns...)}
	for i := range r.columns {
		column := renderLeaf(s, t.fields[r.columns[i].Name], sc)
		r.columns[i].Value, r.columns[i].Null = column.text, column.null
	}
	return r
}

// recordColumns is the sorted non-reference field names — the columns a record
// projects. A {/path} binding is carried in fields under its root path, so only a
// name that is not a reference is a column.
func recordColumns(t *template) []string {
	var names []string
	for _, name := range sortedNames(t.fields) {
		if !isRef(name) {
			names = append(names, name)
		}
	}
	return names
}
