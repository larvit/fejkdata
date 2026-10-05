package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/rows"
)

// table is a category whose rows come from a TSV beside it: the header names the
// columns, each row is one draw, and the format renders the drawn row. A column is a
// cell of the row a path pinned; a cell carrying tokens compiles to a string node.
type table struct {
	rows           *rowsTable
	formatTemplate *template // fields are the column nodes
	rowNode        *tableRow
	cellTemplates  map[int]*template
}

type (
	rowsTable = rows.Table[*table]
	pinSet    = rows.Pins[*table]
)

func (*table) isNode() {}

// tableColumn is one column of a table, rendered as the cell of the row the render pinned.
type tableColumn struct {
	t *table
	i int
}

func (*tableColumn) isNode() {}

func (c *tableColumn) name() string { return c.t.rows.Header()[c.i] }

// tableRow is the row of a table the render pinned, rendered by the table's format.
type tableRow struct{ t *table }

func (*tableRow) isNode() {}

// cellTemplate is what a cell renders: its compiled template where it carries tokens,
// else nil for its text.
func (t *table) cellTemplate(row, col int) *template {
	return t.cellTemplates[row*len(t.rows.Header())+col]
}

// tableOptions are the keys a table object takes; every other key is refused.
var tableOptions = []string{"format", "key", "name", "parent", "rows", "weight"}

func isTableOption(name string) bool {
	for _, o := range tableOptions {
		if o == name {
			return true
		}
	}
	return false
}

func compileTable(m map[string]any, dir []string, name string, readRows func(file string) (string, error)) (*table, error) {
	o, err := readTableOptions(m)
	if err != nil {
		return nil, err
	}
	data, err := readRows(o.rows)
	if err != nil {
		return nil, err
	}
	t := &table{}
	if t.rows, err = rows.Parse(t, name, categoryPath(dir, name), o.rows, data, o.columns); err != nil {
		return nil, err
	}
	if err := t.checkCells(); err != nil {
		return nil, fmt.Errorf("%s: %w", o.rows, err)
	}
	if err := t.compileRowFormat(o.format); err != nil {
		return nil, err
	}
	return t, nil
}

type tableOptionValues struct {
	format, rows string
	columns      rows.Columns
}

func readTableOptions(m map[string]any) (tableOptionValues, error) {
	var o tableOptionValues
	for k := range m {
		if !isTableOption(k) {
			return o, fmt.Errorf("a table takes %s; %q is none of them", strings.Join(tableOptions, ", "), k)
		}
	}
	for k, into := range map[string]*string{"format": &o.format, "key": &o.columns.Key, "name": &o.columns.Name, "parent": &o.columns.Parent, "rows": &o.rows, "weight": &o.columns.Weight} {
		v, ok := m[k]
		if !ok {
			continue
		}
		s, isString := v.(string)
		if !isString {
			return o, fmt.Errorf("%s must be a string, got %T", k, v)
		}
		if k != "format" && s == "" {
			return o, fmt.Errorf("%s names a column, so it cannot be empty; drop it", k)
		}
		*into = s
	}
	if _, ok := m["format"]; !ok {
		return o, fmt.Errorf("template object missing string \"format\"")
	}
	if !strings.HasSuffix(o.rows, ".tsv") || strings.Contains(o.rows, "/") {
		return o, fmt.Errorf("rows names a .tsv file beside the category, not %q", o.rows)
	}
	if o.columns.Key != "" && o.columns.Key == o.columns.Name {
		return o, fmt.Errorf("name names the key column %q, which a selector already reads; drop it", o.columns.Name)
	}
	return o, nil
}

// checkCells compiles every cell carrying a token.
func (t *table) checkCells() error {
	header := t.rows.Header()
	for row := 0; row < t.rows.Len(); row++ {
		for col := range header {
			cell := t.rows.Cell(row, col)
			if strings.IndexByte(cell, '{') < 0 && strings.IndexByte(cell, '}') < 0 {
				continue
			}
			n, err := compileCell(cell)
			if err != nil {
				return fmt.Errorf("line %d, %s: %w", row+2, header[col], err)
			}
			if t.cellTemplates == nil {
				t.cellTemplates = map[int]*template{}
			}
			t.cellTemplates[row*len(header)+col] = n
		}
	}
	return nil
}

// compileCell compiles a cell carrying a token. A cell sits in no name scope, so it binds no
// name and reads only references.
func compileCell(cell string) (*template, error) {
	n, err := compileString(cell)
	if err != nil {
		return nil, err
	}
	if err := refuseTableBinding(n.tokens); err != nil {
		return nil, err
	}
	if len(n.unbound) > 0 {
		return nil, n.unbound[0].err
	}
	return n, nil
}

// compileRowFormat compiles the format a row renders through, over the columns as its
// fields.
func (t *table) compileRowFormat(format string) error {
	toks, err := grammar.ParseFormat(format)
	if err != nil {
		return err
	}
	if err := refuseTableBinding(toks); err != nil {
		return err
	}
	for _, tok := range toks {
		if tok.Kind != grammar.NameRead {
			continue
		}
		for _, name := range tok.Arms {
			a := splitArm(name, nil)
			if _, ok := t.rows.Column(a.head); !ok && !grammar.IsRef(a.head) && a.head != "" {
				return fmt.Errorf("format names no column %q of %s; the columns are %v", a.head, t.rows.File(), t.rows.Header())
			}
		}
	}
	fields := make(map[string]node, len(t.rows.Header()))
	for i, name := range t.rows.Header() {
		fields[name] = &tableColumn{t, i}
	}
	unbound, err := checkTokens(toks, fields)
	if err != nil {
		return err
	}
	if len(unbound) > 0 {
		return unbound[0].err
	}
	t.formatTemplate = &template{format: format, tokens: toks, fields: fields, repeat: 1, isRecord: true}
	t.rowNode = &tableRow{t}
	return nil
}

// refuseTableBinding refuses a table's format or cell binding a name.
func refuseTableBinding(toks []grammar.Token) error {
	for _, tok := range toks {
		if tok.Kind == grammar.NameBind {
			return fmt.Errorf("token {%s}: a table binds no name; bind it in a template reading the table", tok.Body)
		}
	}
	return nil
}

// linkTables links each table to the parent beside it.
func linkTables(sites []categorySite) error {
	for _, s := range sites {
		t, isTable := s.n.(*table)
		if !isTable || t.rows.LinkColumn() < 0 {
			continue
		}
		if err := t.linkParent(s.in.children); err != nil {
			return fmt.Errorf("%s: %w", s.path, err)
		}
	}
	return nil
}

// linkParent links t to the table among siblings that its link column names.
func (t *table) linkParent(siblings map[string]node) error {
	name := t.rows.Options().Parent
	p, ok := siblings[name].(*table)
	switch {
	case siblings[name] == nil:
		return fmt.Errorf("parent %q names no table beside it", name)
	case !ok:
		return fmt.Errorf("parent %q is not a table; a link column reads a table's key", name)
	}
	return t.rows.Link(p.rows, func(name string) *rowsTable {
		if q, ok := siblings[name].(*table); ok {
			return q.rows
		}
		return nil
	})
}
