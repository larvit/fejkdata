package fejkdata

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Field is one rendered column of a record.
type Field struct {
	Name  string
	Value string
}

// Record is one record rendered from a template: every direct field is a column,
// drawn independently and listed in name order. The template's format is the
// string a [Generator.Fake] call renders; a record is its inverse — each field
// projected as a column instead of composed.
type Record struct {
	fields []Field
}

// Fields returns the record's columns in name order.
func (r *Record) Fields() []Field {
	return append([]Field(nil), r.fields...)
}

// JSON renders the record as one JSON object, every column a string.
func (r *Record) JSON() string {
	m := make(map[string]string, len(r.fields))
	for _, f := range r.fields {
		m[f.Name] = f.Value
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
	out := make([]string, len(r.fields))
	for i, f := range r.fields {
		out[i] = f.Name
	}
	return out
}

func (r *Record) values() []string {
	out := make([]string, len(r.fields))
	for i, f := range r.fields {
		out[i] = f.Value
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

// SQLInsert renders the record as one INSERT statement into table, every column a
// single-quoted string literal.
func (r *Record) SQLInsert(table string) string {
	cols := make([]string, len(r.fields))
	vals := make([]string, len(r.fields))
	for i, f := range r.fields {
		cols[i] = f.Name
		vals[i] = "'" + strings.ReplaceAll(f.Value, "'", "''") + "'"
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", table, strings.Join(cols, ", "), strings.Join(vals, ", "))
}

// Record renders a path as one record: the template it names, with each direct
// field drawn as a column. A path naming a folder or a value that is not a
// template — a bare string or a choice, which have no fields — is an error.
func (f *Generator) Record(path string) (*Record, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, err := descend(f.rand, &group{children: f.categories}, strings.Split(path, "."))
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	if _, ok := n.(*group); ok {
		return nil, fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	t, ok := n.(*template)
	if !ok {
		return nil, fmt.Errorf("fejkdata: %s does not name a record; it has no fields to project", path)
	}
	r := renderRecord(f.rand, t)
	if len(r.fields) == 0 {
		return nil, fmt.Errorf("fejkdata: %s has no fields, so no columns", path)
	}
	return r, nil
}

// RecordTemplate is an inline record compiled, referenced and validated once,
// ready to render many times with [RecordTemplate.Fake].
type RecordTemplate struct {
	g *Generator
	n node
}

// Fake renders the record with one draw.
func (t *RecordTemplate) Fake() *Record {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return renderRecord(t.g.rand, t.n.(*template))
}

// NewRecordTemplate compiles an inline record — a JSON object with a format and
// fields — and binds its references against the loaded tree.
func (f *Generator) NewRecordTemplate(input string) (*RecordTemplate, error) {
	n, err := compileInput(input)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	t, ok := n.(*template)
	if !ok {
		return nil, fmt.Errorf("fejkdata: an inline record is a JSON object with a format and fields, not a choice")
	}
	if len(recordColumns(t)) == 0 {
		return nil, fmt.Errorf("fejkdata: an inline record needs at least one field to project as a column")
	}
	scope := inlineScope(t)
	if err := linkNodeRefs(scope, f.categories); err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	if err := checkScope(scope); err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &RecordTemplate{g: f, n: t}, nil
}

// FakeRecord compiles and renders an inline record in one call.
func (f *Generator) FakeRecord(input string) (*Record, error) {
	t, err := f.NewRecordTemplate(input)
	if err != nil {
		return nil, err
	}
	return t.Fake(), nil
}

// renderRecord projects a template's direct fields as columns, drawn once each,
// in name order. A {/path} binding is a render edge, not a column, so it is
// skipped the same way List and the graph do.
func renderRecord(s *session, t *template) *Record {
	r := &Record{}
	for _, name := range recordColumns(t) {
		r.fields = append(r.fields, Field{Name: name, Value: render(s, t.fields[name])})
	}
	return r
}

// recordColumns is the sorted non-reference field names — the columns a record
// projects. A {/path} binding is carried in fields under its root path, so only a
// name that is not a reference is a column.
func recordColumns(t *template) []string {
	var names []string
	for name := range t.fields {
		if !isRef(name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
