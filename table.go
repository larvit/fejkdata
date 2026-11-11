package fejkdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/jsonvalue"
	"github.com/larvit/fejkdata/internal/rows"
)

var tableOptions = []string{"format", "key", "name", "parent", "rows", "weight"}

func compileTable(m map[string]any, dir []string, name string, readRows func(file string) (string, error)) (*tree.Table, error) {
	o, err := readTableOptions(m)
	if err != nil {
		return nil, err
	}
	data, err := readRows(o.rows)
	if err != nil {
		return nil, err
	}
	t := &tree.Table{}
	rs, err := rows.Parse(t, name, categoryPath(dir, name), o.rows, data, o.options)
	if err != nil {
		return nil, err
	}
	t.SetRows(rs)
	cells, err := compileCells(rs)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", o.rows, err)
	}
	format, err := compileRowFormat(t, rs, o.format)
	if err != nil {
		return nil, err
	}
	t.SetSyntax(format, cells)
	return t, nil
}

type tableOptionValues struct {
	format, rows string
	options      rows.Options
}

func readTableOptions(m map[string]any) (tableOptionValues, error) {
	var o tableOptionValues
	for k := range m {
		if !slices.Contains(tableOptions, k) {
			return o, fmt.Errorf("a table takes %s; %q is none of them", strings.Join(tableOptions, ", "), k)
		}
	}
	for k, into := range map[string]*string{"format": &o.format, "key": &o.options.Key, "name": &o.options.Name, "parent": &o.options.Parent, "rows": &o.rows, "weight": &o.options.Weight} {
		v, ok := m[k]
		if !ok {
			continue
		}
		s, isString := v.(string)
		if !isString {
			return o, fmt.Errorf("%s must be a string, not %s", k, jsonvalue.Kind(v))
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
	return o, nil
}

// compileCells compiles each cell of rs that carries tokens, by cellIndex.
func compileCells(rs *tree.RowsTable) (map[int]*tree.Template, error) {
	var cells map[int]*tree.Template
	header := rs.Header()
	for row := 0; row < rs.Len(); row++ {
		for col := range header {
			cell := rs.Cell(row, col)
			if strings.IndexByte(cell, '{') < 0 && strings.IndexByte(cell, '}') < 0 {
				continue
			}
			n, err := compileCell(cell)
			if err != nil {
				return nil, fmt.Errorf("line %d, %s: %w", row+2, header[col], err)
			}
			if cells == nil {
				cells = map[int]*tree.Template{}
			}
			cells[tree.CellIndex(len(header), row, col)] = n
		}
	}
	return cells, nil
}

// compileCell compiles a cell carrying a token. A cell sits in no name scope, so it binds no
// name and reads only references.
func compileCell(cell string) (*tree.Template, error) {
	n, err := compileString(cell)
	if err != nil {
		return nil, err
	}
	if err := refuseInTable(n.Tokens()); err != nil {
		return nil, err
	}
	if len(n.Unbound()) > 0 {
		return nil, n.Unbound()[0].Err()
	}
	return n, nil
}

// compileRowFormat compiles the format a row renders through, over the columns as its
// fields.
func compileRowFormat(t *tree.Table, rs *tree.RowsTable, format string) (*tree.Template, error) {
	toks, err := grammar.ParseFormat(format)
	if err != nil {
		return nil, err
	}
	if err := refuseInTable(toks); err != nil {
		return nil, err
	}
	for _, tok := range toks {
		if tok.Kind != grammar.PathRead {
			continue
		}
		for _, name := range tok.Arms {
			a := splitRead(name, nil)
			if _, ok := rs.Column(a.head); !ok && !grammar.IsRef(a.head) && a.head != "" {
				return nil, fmt.Errorf("format names no column %q of %s; the columns are %v", a.head, rs.File(), rs.Header())
			}
		}
	}
	fields := make(map[string]tree.Node, len(rs.Header()))
	for i, name := range rs.Header() {
		fields[name] = tree.NewTableColumn(t, i)
	}
	unbound, err := checkTokens(toks, fields)
	if err != nil {
		return nil, err
	}
	if len(unbound) > 0 {
		return nil, unbound[0].Err()
	}
	return tree.NewTemplate(tree.TemplateSyntax{Format: format, Tokens: toks, Fields: fields, Repeat: 1, IsRecord: true}), nil
}

// refuseInTable refuses a table's format or cell binding a name, filtering a draw or drawing
// distinct values.
func refuseInTable(toks []grammar.Token) error {
	for _, tok := range toks {
		switch {
		case tok.Kind == grammar.NameBind:
			return fmt.Errorf("token {%s}: a table binds no name; bind it in a template reading the table", tok.Body)
		case tok.Where != nil:
			return fmt.Errorf("token {%s}: a table filters no draw; filter it in a template reading the table", tok.Body)
		case tok.Distinct != nil:
			return fmt.Errorf("token {%s}: a table draws no distinct values; draw them in a template reading the table", tok.Body)
		}
	}
	return nil
}

func linkTables(sites []categorySite) error {
	for _, s := range sites {
		if t, isTable := s.n.(*tree.Table); isTable {
			if err := linkTable(t, s.in.Children()); err != nil {
				return fmt.Errorf("%s: %w", s.path, err)
			}
		}
	}
	return nil
}

// linkTable gives t the folder holding it, siblings, and links it to its parent there.
func linkTable(t *tree.Table, siblings map[string]tree.Node) error {
	t.SetSiblings(siblings)
	if t.Rows().Options().Parent == "" {
		return nil
	}
	return linkParent(t, siblings)
}

// linkParent links t to the table among siblings that its link column names.
func linkParent(t *tree.Table, siblings map[string]tree.Node) error {
	name := t.Rows().Options().Parent
	p, ok := siblings[name].(*tree.Table)
	switch {
	case siblings[name] == nil:
		return fmt.Errorf("parent %q names no table beside it", name)
	case !ok:
		return fmt.Errorf("parent %q is not a table; a link column reads a table's key", name)
	}
	return t.Rows().Link(p.Rows(), func(sibling string) *tree.RowsTable {
		if q, ok := siblings[sibling].(*tree.Table); ok {
			return q.Rows()
		}
		return nil
	})
}

// tableRoute is how a path passes one table: the row the path reads there, by selector or
// drawn, and where the walk stands past the table.
type tableRoute struct {
	sel  string
	draw bool
	then pathPos
	up   bool
}

// routeThrough is how tail passes t. fromRow means the previous route stepped into t
// from a row of a table linked to it, so t reads a row even where tail is empty; a
// table reached otherwise, with no selector and an empty tail, is left to a render's
// own draw.
func routeThrough(t *tree.Table, tail []string, fromRow bool) (tableRoute, error) {
	sel, tail, err := splitSelector(t, tail)
	if err != nil {
		return tableRoute{}, err
	}
	if len(tail) > 0 && tail[0] == ".." {
		if err := t.Rows().ProveStepUp(tail[1:]); err != nil {
			return tableRoute{}, err
		}
		return tableRoute{sel: sel, draw: sel == "", then: pathPos{n: t.Rows().Parent().Owner(), tail: tail[2:], fromRow: true}, up: true}, nil
	}
	column, child, err := columnOrLinked(t, tail)
	if err != nil {
		return tableRoute{}, err
	}
	selected := sel != ""
	readsRow := selected || fromRow || len(tail) > 0
	// A selector further down pins this table by ancestry; drawing first could draw a row the
	// selector is not inside.
	pinnedBelow := grammar.HasSelector(tail)
	r := tableRoute{sel: sel, draw: !selected && readsRow && !pinnedBelow}
	switch {
	case !readsRow:
		r.then = pathPos{n: t}
	case len(tail) == 0:
		r.then = pathPos{n: t.RowNode()}
	case child != nil:
		r.then = pathPos{n: child, tail: tail[1:], fromRow: true}
	default:
		r.then = pathPos{n: column, tail: tail[1:]}
	}
	return r, nil
}

// routeSteps appends the steps r takes past t, the first at at, pinning in pins the row
// it selects.
func routeSteps(steps []tree.PathStep, pins *tree.PinSet, t *tree.Table, r tableRoute, at int) ([]tree.PathStep, error) {
	if r.sel != "" {
		row, err := pins.Select(t.Rows(), r.sel)
		if err != nil {
			return steps, err
		}
		steps = append(steps, tree.PathStep{Kind: tree.StepSelect, At: at, Name: r.sel, Row: row})
		at++
	}
	if r.draw {
		steps = append(steps, tree.PathStep{Kind: tree.StepDraw, At: at})
	}
	switch next := r.then.n.(type) {
	case *tree.Table:
		switch {
		case r.up:
			steps = append(steps, tree.PathStep{Kind: tree.StepParent, At: at})
		case next != t:
			steps = append(steps, tree.PathStep{Kind: tree.StepChild, At: at, Name: next.Rows().Segment()})
		}
	case *tree.TableRow:
		steps = append(steps, tree.PathStep{Kind: tree.StepRow, At: at})
	case *tree.TableColumn:
		steps = append(steps, tree.PathStep{Kind: tree.StepColumn, At: at, Name: next.Name()})
	default:
		panic(invariant.Broken("%s: a route onto %T", t.Rows().Segment(), r.then.n))
	}
	return steps, nil
}

func followStep(t *tree.Table, s *drawstate.State, st tree.PathStep, pins *tree.PinSet) tree.Node {
	switch st.Kind {
	case tree.StepSelect:
		if err := pins.PinRow(t.Rows(), st.Row); err != nil {
			panic(invariant.Broken("%s[%s]: %v", t.Rows().Segment(), st.Name, err))
		}
		return t
	case tree.StepDraw:
		t.Rows().DrawIn(s, pins)
		return t
	case tree.StepRow:
		return t.RowNode()
	case tree.StepColumn:
		return t.FormatTemplate().Fields()[st.Name]
	case tree.StepChild:
		return t.Descendant(st.Name)
	case tree.StepParent:
		return t.Rows().Parent().Owner()
	}
	panic(invariant.Broken("followStep has no case for step kind %d", st.Kind))
}

func splitSelector(t *tree.Table, tail []string) (sel string, rest []string, err error) {
	if len(tail) == 0 || !grammar.IsSelector(tail[0]) {
		return "", tail, nil
	}
	sel, rest = grammar.SelectorOf(tail[0]), tail[1:]
	if len(rest) > 0 && grammar.IsSelector(rest[0]) {
		return "", nil, fmt.Errorf("%s[%s] is selected twice; one selector names its row", t.Rows().Segment(), sel)
	}
	return sel, rest, nil
}

func columnOrLinked(t *tree.Table, tail []string) (column tree.Node, child *tree.Table, err error) {
	if len(tail) == 0 {
		return nil, nil, nil
	}
	if column, ok := t.FormatTemplate().Fields()[tail[0]]; ok {
		return column, nil, nil
	}
	d := t.Descendant(tail[0])
	if d == nil {
		return nil, nil, fmt.Errorf("no column or linked table %q in %s", tail[0], t.Rows().Segment())
	}
	return nil, d, nil
}
