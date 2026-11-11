package tree

import (
	"slices"

	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/rows"
)

// Table is a category whose rows come from a TSV beside it: the header names the
// columns, each row is one draw, and the format renders the drawn row. A column is a
// cell of the row a path pinned; a cell carrying tokens compiles to a string node.
type Table struct {
	rows           *RowsTable
	formatTemplate *Template // fields are the column nodes
	rowNode        *TableRow
	cellTemplates  map[int]*Template // by CellIndex
	siblings       map[string]Node   // the folder holding t, set by SetSiblings; every table linked to t sits there
}

func (t *Table) CellTemplates() map[int]*Template { return t.cellTemplates }

func (t *Table) FormatTemplate() *Template { return t.formatTemplate }

func (t *Table) RowNode() *TableRow { return t.rowNode }

func (t *Table) Rows() *RowsTable { return t.rows }

type (
	PinSet    = rows.Pins[*Table]
	RowsTable = rows.Table[*Table]
)

func (*Table) isNode() {}

// TableColumn is one column of a table, rendered as the cell of the row the render pinned.
type TableColumn struct {
	t *Table
	i int
}

func (*TableColumn) isNode() {}

func (c *TableColumn) Name() string { return c.t.rows.Header()[c.i] }

func (c *TableColumn) Index() int { return c.i }

func (c *TableColumn) Table() *Table { return c.t }

// TableRow is the row of a table the render pinned, rendered by the table's format.
type TableRow struct{ t *Table }

func NewTableColumn(t *Table, i int) *TableColumn { return &TableColumn{t, i} }

func (*TableRow) isNode() {}

func (t *TableRow) Table() *Table { return t.t }

// CellTemplate is what a cell renders: its compiled template where it carries tokens,
// else nil for its text.
func (t *Table) CellTemplate(row, col int) *Template {
	return t.cellTemplates[CellIndex(len(t.rows.Header()), row, col)]
}

// CellIndex keys the cell at row and col of a table width columns wide.
func CellIndex(width, row, col int) int { return row*width + col }

// SetRows gives t its rows, parsed with t as their owner, before its format compiles against them.
func (t *Table) SetRows(rs *RowsTable) {
	if t.rows != nil {
		panic(invariant.Broken("table %s is parsed twice", rs.File()))
	}
	t.rows = rs
}

// SetSyntax gives t what its compile built from its rows: the format a row renders through, and
// each cell carrying tokens, by CellIndex.
func (t *Table) SetSyntax(format *Template, cells map[int]*Template) {
	if t.formatTemplate != nil {
		panic(invariant.Broken("table %s is compiled twice", t.rows.File()))
	}
	t.formatTemplate, t.rowNode, t.cellTemplates = format, &TableRow{t}, cells
}

// SetSiblings gives t the folder holding it; every table linked to t sits there.
func (t *Table) SetSiblings(siblings map[string]Node) { t.siblings = siblings }

// ChildTables is each table whose parent is t, in name order.
func (t *Table) ChildTables() []*Table {
	names := make([]string, 0, len(t.siblings))
	for name := range t.siblings {
		names = append(names, name)
	}
	slices.Sort(names)
	var out []*Table
	for _, name := range names {
		if c, isTable := t.siblings[name].(*Table); isTable && c.rows.Parent() == t.rows {
			out = append(out, c)
		}
	}
	return out
}

// Descendant is the table named name among those linked to t, at any depth, or nil.
func (t *Table) Descendant(name string) *Table {
	d, isTable := t.siblings[name].(*Table)
	if !isTable {
		return nil
	}
	for p := d.rows.Parent(); p != nil; p = p.Parent() {
		if p == t.rows {
			return d
		}
	}
	return nil
}
