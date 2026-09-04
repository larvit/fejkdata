package fejkdata

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Column is one rendered column of a record.
type Column struct {
	Name  string
	Value string
}

// Record is one record rendered from a template: every direct field is a column,
// drawn independently and listed in name order. The template's format is the
// string a [Generator.Fake] call renders; a record is its inverse — each field
// projected as a column instead of composed.
type Record struct {
	columns []Column
}

// Columns returns the record's columns in name order.
func (r *Record) Columns() []Column {
	return append([]Column(nil), r.columns...)
}

// JSON renders the record as one JSON object, every column a string.
func (r *Record) JSON() string {
	m := make(map[string]string, len(r.columns))
	for _, c := range r.columns {
		m[c.Name] = c.Value
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// CSVHeader renders the column names as one CSV header line.
func (r *Record) CSVHeader() string {
	return csvLine(r.names())
}

// CSVLine renders the column values as one CSV row.
func (r *Record) CSVLine() string {
	return csvLine(r.values())
}

func (r *Record) names() []string {
	out := make([]string, len(r.columns))
	for i, c := range r.columns {
		out[i] = c.Name
	}
	return out
}

func (r *Record) values() []string {
	out := make([]string, len(r.columns))
	for i, c := range r.columns {
		out[i] = c.Value
	}
	return out
}

func csvLine(cols []string) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write(cols)
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}

// SQLInsert renders the record as one INSERT statement into table: identifiers in
// ANSI double quotes, every value a single-quoted string literal.
func (r *Record) SQLInsert(table string) string {
	cols := make([]string, len(r.columns))
	vals := make([]string, len(r.columns))
	for i, c := range r.columns {
		cols[i] = quoteIdent(c.Name)
		vals[i] = "'" + strings.ReplaceAll(c.Value, "'", "''") + "'"
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", quoteIdent(table), strings.Join(cols, ", "), strings.Join(vals, ", "))
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// Record renders a path as one record: the template it names, with each direct
// field drawn as a column. Only a category-level template is a record — a path
// that descends into a field, or that names a folder or a choice, is an error.
func (f *Generator) Record(path string) (*Record, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, n, tail, err := resolveCategory(f.categories, strings.Split(path, "."))
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	if len(tail) > 0 {
		return nil, fmt.Errorf("fejkdata: %s descends into %q, a field; only a category-level template is a record", path, tail[0])
	}
	t, columns, err := recordOf(n)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %s %w", path, err)
	}
	return renderRecord(f.rand, t, columns), nil
}

// RecordTemplate is an inline record compiled, referenced and validated once,
// ready to render many times with [RecordTemplate.Fake].
type RecordTemplate struct {
	g       *Generator
	t       *template
	columns []string
}

// Fake renders the record with one draw.
func (t *RecordTemplate) Fake() *Record {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return renderRecord(t.g.rand, t.t, t.columns)
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

// FakeRecord compiles and renders an inline record in one call.
func (f *Generator) FakeRecord(input string) (*Record, error) {
	t, err := f.NewRecordTemplate(input)
	if err != nil {
		return nil, err
	}
	return t.Fake(), nil
}

// recordOf is the fence both record entry points pass: the node is a template, it
// carries no repeat — which composes the format rather than projecting columns —
// and it offers at least one column. The columns come back with it, fixed for
// every draw the caller goes on to make.
func recordOf(n node) (*template, []string, error) {
	t, ok := n.(*template)
	if !ok {
		return nil, nil, errors.New("names a choice, not a template; a record is a template whose fields are its columns")
	}
	if t.repeat != 1 {
		return nil, nil, fmt.Errorf("carries repeat %d, which composes its format into one string; a record projects columns instead — drop the repeat and render the record again for more rows", t.repeat)
	}
	columns := recordColumns(t)
	if len(columns) == 0 {
		return nil, nil, errors.New("has no fields, so no columns")
	}
	return t, columns, nil
}

// renderRecord draws each column once, in the name order recordOf fixed. The
// columns share one reference scope, so two columns that reference one category
// read one draw of it.
func renderRecord(s *session, t *template, columns []string) *Record {
	scope := &draws{variant: map[string]node{}, value: map[string]string{}}
	r := &Record{columns: make([]Column, len(columns))}
	for i, name := range columns {
		r.columns[i] = Column{Name: name, Value: render(s, t.fields[name], scope)}
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
