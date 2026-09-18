package fejkdata

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// table is a category whose rows come from a TSV beside it: the header names the
// columns, each row is one draw, and the format renders the drawn row. A column is a
// cell of the row a render pinned; a cell carrying tokens compiles to a string node.
type table struct {
	category string
	file     string
	format   *template // fields are the column nodes
	columns  []string
	col      map[string]int
	fields   map[string]node // column nodes, the format's fields
	whole    *row            // the pinned row rendered by the format
	cells    []string        // rows × columns, flat
	tokens   map[int]*template
	key      int // column index, or -1
	name     int
	weight   int
	parent   int
	cum      []float64 // cumulative weights, nil when uniform
	byKey    map[string]int
	parentT  *table
	children map[string]*table
	once     sync.Once
	index    tableIndex
}

// tableIndex is what selection and linked draws look up, built on the first draw
// that needs it.
type tableIndex struct {
	byName   map[string][]int
	children map[string][]int     // rows by parent key
	childCum map[string][]float64 // cumulative weights per parent key, nil when uniform
}

func (*table) isNode() {}

// column is one column of a table, rendered as the cell of the row the render pinned.
type column struct {
	t *table
	i int
}

func (*column) isNode() {}

// row is the row of a table the render pinned, rendered by the table's format.
type row struct{ t *table }

func (*row) isNode() {}

func (t *table) rows() int { return len(t.cells) / len(t.columns) }

func (t *table) cell(row, col int) string { return t.cells[row*len(t.columns)+col] }

// cellNode is what a cell renders: its compiled template where it carries tokens,
// else nil for its text.
func (t *table) cellNode(row, col int) *template { return t.tokens[row*len(t.columns)+col] }

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

// inSelector is what a key or name cell may not contain: the path grammar reserves
// brackets and braces, and a quote or a bar would make the selector no path and no
// token arm.
const inSelector = `[]{}"|`

// compileTable compiles a category object naming a rows file into a table.
func compileTable(m map[string]any, category string, files *categoryFiles) (*table, error) {
	o, err := readTableOptions(m)
	if err != nil {
		return nil, err
	}
	data, err := files.rows(o.rows)
	if err != nil {
		return nil, err
	}
	t := &table{category: category, file: o.rows, key: -1, name: -1, weight: -1, parent: -1}
	if err := t.parseRows(data); err != nil {
		return nil, fmt.Errorf("%s: %w", o.rows, err)
	}
	if err := t.bindOptions(o); err != nil {
		return nil, err
	}
	if err := t.checkCells(); err != nil {
		return nil, fmt.Errorf("%s: %w", o.rows, err)
	}
	if err := t.compileFormat(o.format); err != nil {
		return nil, err
	}
	return t, nil
}

type tableOptionValues struct {
	format, key, name, parent, rows, weight string
}

func readTableOptions(m map[string]any) (tableOptionValues, error) {
	var o tableOptionValues
	for k := range m {
		if !isTableOption(k) {
			return o, fmt.Errorf("a table takes %s; %q is none of them", strings.Join(tableOptions, ", "), k)
		}
	}
	for k, into := range map[string]*string{"format": &o.format, "key": &o.key, "name": &o.name, "parent": &o.parent, "rows": &o.rows, "weight": &o.weight} {
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
	if o.key != "" && o.key == o.name {
		return o, fmt.Errorf("name names the key column %q, which a selector already reads; drop it", o.name)
	}
	return o, nil
}

// parseRows reads the TSV: the header line names the columns, each following line
// is a row of as many cells. Cells are substrings of data, so the file is held once.
func (t *table) parseRows(data string) error {
	data = strings.TrimSuffix(strings.TrimPrefix(data, "\xEF\xBB\xBF"), "\n")
	header, rest, _ := strings.Cut(data, "\n")
	header = strings.TrimSuffix(header, "\r")
	if header == "" {
		return fmt.Errorf("has no header line naming its columns")
	}
	t.columns = strings.Split(header, "\t")
	t.col = make(map[string]int, len(t.columns))
	t.fields = make(map[string]node, len(t.columns))
	for i, name := range t.columns {
		if err := checkName(name); err != nil {
			return fmt.Errorf("column %w", err)
		}
		if _, dup := t.col[name]; dup {
			return fmt.Errorf("column %q is named twice", name)
		}
		t.col[name] = i
		t.fields[name] = &column{t, i}
	}
	if header == data {
		return fmt.Errorf("has no rows below its header; a table is at least two rows")
	}
	m := len(t.columns)
	t.cells = make([]string, 0, m*(strings.Count(rest, "\n")+1))
	for line := 2; ; line++ {
		row, more, found := strings.Cut(rest, "\n")
		row = strings.TrimSuffix(row, "\r")
		if err := t.appendRow(row, line); err != nil {
			return err
		}
		if !found {
			break
		}
		rest = more
	}
	if t.rows() < 2 {
		return fmt.Errorf("has one row, which is a template; write it as one")
	}
	return nil
}

// appendRow splits one line into as many cells as the header has columns.
func (t *table) appendRow(row string, line int) error {
	if row == "" {
		return fmt.Errorf("line %d is empty; every line below the header is a row", line)
	}
	m := len(t.columns)
	for n := 0; n < m; n++ {
		cell, next, tab := strings.Cut(row, "\t")
		if !tab && n < m-1 || tab && n == m-1 {
			return fmt.Errorf("line %d has %s cells than the %d columns", line, map[bool]string{true: "more", false: "fewer"}[tab], m)
		}
		t.cells = append(t.cells, cell)
		row = next
	}
	return nil
}

// bindOptions resolves each option to its column and proves what it claims of the
// cells: a key is unique, a weight a positive number.
func (t *table) bindOptions(o tableOptionValues) error {
	for _, opt := range []struct {
		name, value string
		into        *int
	}{{"key", o.key, &t.key}, {"name", o.name, &t.name}, {"parent", o.parent, &t.parent}, {"weight", o.weight, &t.weight}} {
		if opt.value == "" {
			continue
		}
		i, ok := t.col[opt.value]
		if !ok {
			return fmt.Errorf("%s names no column %q of %s; the columns are %v", opt.name, opt.value, o.rows, t.columns)
		}
		*opt.into = i
	}
	if t.name >= 0 && t.key < 0 && t.parent < 0 {
		return fmt.Errorf("name selects a row as a key does, and lists the rows it matches by their keys, so it needs a key column, or a parent inside which each name is one row; add key")
	}
	if err := t.indexKeys(); err != nil {
		return err
	}
	if err := t.proveNamesInsideParent(); err != nil {
		return err
	}
	return t.sumWeights()
}

// proveNamesInsideParent proves a name without a key names one row inside its parent.
func (t *table) proveNamesInsideParent() error {
	if t.key >= 0 || t.name < 0 {
		return nil
	}
	inside := make(map[string]int, t.rows())
	for r := 0; r < t.rows(); r++ {
		k := t.cell(r, t.parent) + "\t" + t.cell(r, t.name)
		if first, dup := inside[k]; dup {
			return fmt.Errorf("%s line %d: name %q repeats line %d inside %s %q; a name selects one row inside its parent", t.file, r+2, t.cell(r, t.name), first+2, t.columns[t.parent], t.cell(r, t.parent))
		}
		inside[k] = r
	}
	return nil
}

// indexKeys proves every key names one row, and keeps the index a link is proved by.
func (t *table) indexKeys() error {
	if t.key < 0 {
		return nil
	}
	t.byKey = make(map[string]int, t.rows())
	for r := 0; r < t.rows(); r++ {
		k := t.cell(r, t.key)
		if k == "" {
			return fmt.Errorf("%s line %d: the key is empty", t.file, r+2)
		}
		if first, dup := t.byKey[k]; dup {
			return fmt.Errorf("%s line %d: key %q repeats line %d; a key selects one row", t.file, r+2, k, first+2)
		}
		t.byKey[k] = r
	}
	for r := 0; r < t.rows() && t.name >= 0; r++ {
		if n := t.cell(r, t.name); t.byKey[n] != r {
			if other, isKey := t.byKey[n]; isKey {
				return fmt.Errorf("%s line %d: name %q is the key of line %d, which a selector reads first, so the name could never select this row", t.file, r+2, n, other+2)
			}
		}
	}
	return nil
}

// sumWeights proves every weight is a positive number and builds the cumulative table.
func (t *table) sumWeights() error {
	if t.weight < 0 {
		return nil
	}
	t.cum = make([]float64, t.rows())
	total := 0.0
	for r := range t.cum {
		w, err := strconv.ParseFloat(t.cell(r, t.weight), 64)
		if err != nil || math.IsInf(w, 0) || math.IsNaN(w) || w <= 0 {
			return fmt.Errorf("%s line %d: weight %q is not a positive number", t.file, r+2, t.cell(r, t.weight))
		}
		total += w
		t.cum[r] = total
	}
	if math.IsInf(total, 0) {
		return fmt.Errorf("%s: the weights sum past the largest number", t.file)
	}
	return nil
}

// checkCells compiles every cell carrying a token, and fences what a selector reads.
func (t *table) checkCells() error {
	for i, cell := range t.cells {
		row, col := i/len(t.columns), i%len(t.columns)
		if (col == t.key || col == t.name) && strings.ContainsAny(cell, inSelector) {
			return fmt.Errorf("line %d: %s %q contains %q, which a selector cannot spell", row+2, t.columns[col], cell, cell[strings.IndexAny(cell, inSelector):][:1])
		}
		if strings.IndexByte(cell, '{') < 0 && strings.IndexByte(cell, '}') < 0 {
			continue
		}
		n, err := compileString(cell)
		if err != nil {
			return fmt.Errorf("line %d, %s: %w", row+2, t.columns[col], err)
		}
		if t.tokens == nil {
			t.tokens = map[int]*template{}
		}
		t.tokens[i] = n.(*template)
		t.tokens[i].cellOf, t.tokens[i].cellRow = t, row
	}
	return nil
}

// compileFormat compiles the format over the columns as its fields.
func (t *table) compileFormat(format string) error {
	for _, name := range fieldTokens(format) {
		a := splitArm(name, nil)
		if _, ok := t.col[a.key]; !ok && !isRef(a.key) && a.key != "" {
			return fmt.Errorf("format names no column %q of %s; the columns are %v", a.key, t.file, t.columns)
		}
	}
	if err := checkTokens(format, t.fields); err != nil {
		return err
	}
	t.format = &template{format: format, fields: t.fields, repeat: 1, record: true, table: t}
	t.whole = &row{t}
	return t.format.compileFormat()
}

// linkTables binds every table's parent to the table beside it, and proves the
// links: a parent has a key, every link cell is one, every parent row is linked
// to, no chain of parents closes, and no child is named like a parent's column.
func linkTables(root map[string]node) error {
	var walk func(dir string, children map[string]node) error
	walk = func(dir string, children map[string]node) error {
		for _, name := range sortedNames(children) {
			path := join(dir, name)
			switch n := children[name].(type) {
			case *folder:
				if err := walk(path, n.children); err != nil {
					return err
				}
			case *table:
				if n.parent < 0 {
					continue
				}
				if err := n.linkParent(path, children); err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
			}
		}
		return nil
	}
	return walk("", root)
}

func (t *table) linkParent(path string, siblings map[string]node) error {
	name := t.columns[t.parent]
	p, ok := siblings[name].(*table)
	switch {
	case siblings[name] == nil:
		return fmt.Errorf("parent %q names no table beside it", name)
	case !ok:
		return fmt.Errorf("parent %q is not a table; a link column reads a table's key", name)
	case p.key < 0:
		return fmt.Errorf("parent %q has no key column to link to", name)
	}
	var ancestors []*table
	for q, seen := p, map[*table]bool{t: true}; q != nil; q, _ = siblings[q.columns[q.parent]].(*table) {
		if seen[q] {
			return fmt.Errorf("parent cycle: %s reaches itself through its parents", q.category)
		}
		seen[q] = true
		ancestors = append(ancestors, q)
		if q.parent < 0 {
			break
		}
	}
	for _, q := range ancestors {
		if _, clash := q.col[t.category]; clash {
			return fmt.Errorf("%q is named like a column of %q, its ancestor, so %s.%s could read either; rename one", t.category, q.category, q.category, t.category)
		}
	}
	linked := make(map[string]bool, p.rows())
	for r := 0; r < t.rows(); r++ {
		k := t.cell(r, t.parent)
		if _, ok := p.byKey[k]; !ok {
			return fmt.Errorf("%s line %d: %s %q is no key of %s", t.file, r+2, name, k, p.file)
		}
		linked[k] = true
	}
	for r := 0; r < p.rows(); r++ {
		if k := p.cell(r, p.key); !linked[k] {
			return fmt.Errorf("%s links no row to %s %q; every %s row needs one, or drop line %d of %s", t.file, name, k, name, r+2, p.file)
		}
	}
	t.parentT = p
	if p.children == nil {
		p.children = map[string]*table{}
	}
	p.children[t.category] = t
	return nil
}

func (t *table) indexed() *tableIndex {
	t.once.Do(func() {
		if t.name >= 0 {
			t.index.byName = make(map[string][]int, t.rows())
			for r := 0; r < t.rows(); r++ {
				n := t.cell(r, t.name)
				t.index.byName[n] = append(t.index.byName[n], r)
			}
		}
		if t.parent >= 0 {
			t.index.children = map[string][]int{}
			for r := 0; r < t.rows(); r++ {
				k := t.cell(r, t.parent)
				t.index.children[k] = append(t.index.children[k], r)
			}
			if t.cum != nil {
				t.index.childCum = make(map[string][]float64, len(t.index.children))
				for k, rows := range t.index.children {
					cum, total := make([]float64, len(rows)), 0.0
					for i, r := range rows {
						w, _ := strconv.ParseFloat(t.cell(r, t.weight), 64) // sumWeights proved it
						total += w
						cum[i] = total
					}
					t.index.childCum[k] = cum
				}
			}
		}
	})
	return &t.index
}

// draw picks a row over the whole table. The session is concrete rather than the rng
// interface so that a walk holding the draws allocates nothing.
func (t *table) draw(s *session) int {
	if t.cum == nil {
		return s.IntN(t.rows())
	}
	return pickCum(s, t.cum)
}

func pickCum(s *session, cum []float64) int {
	x := s.Float64() * cum[len(cum)-1]
	return min(sort.Search(len(cum), func(i int) bool { return cum[i] > x }), len(cum)-1) // x can round up to the total
}

// drawUnder picks a row among those linked to parent row pr.
func (t *table) drawUnder(s *session, pr int) int {
	ix := t.indexed()
	k := t.parentT.cell(pr, t.parentT.key)
	rows := ix.children[k]
	if ix.childCum == nil {
		return rows[s.IntN(len(rows))]
	}
	return rows[pickCum(s, ix.childCum[k])]
}

// descendant is the table named name among those linked to t, at any depth.
func (t *table) descendant(name string) *table {
	if c, ok := t.children[name]; ok {
		return c
	}
	for _, c := range t.children {
		if d := c.descendant(name); d != nil {
			return d
		}
	}
	return nil
}

// parentRow is the row of t's parent that row r links to.
func (t *table) parentRow(r int) int { return t.parentT.byKey[t.cell(r, t.parent)] }

// descends reports whether a is an ancestor of t.
func (t *table) descends(a *table) bool {
	for p := t.parentT; p != nil; p = p.parentT {
		if p == a {
			return true
		}
	}
	return false
}

// under reports whether row r of t sits inside row pr of ancestor a.
func (t *table) under(r int, a *table, pr int) bool {
	for c, row := t, r; c.parentT != nil; c, row = c.parentT, c.parentRow(row) {
		if c.parentT == a {
			return c.parentRow(row) == pr
		}
	}
	return false
}

// find is the row a selector names: by key first, then by name, where a name
// naming several rows resolves inside the ancestors pinned in d.
func (t *table) find(sel string, d *draws) (int, error) {
	if t.key < 0 && t.name < 0 {
		return 0, fmt.Errorf("%s has no key or name column to select a row by", t.category)
	}
	if r, ok := t.byKey[sel]; ok {
		return r, nil
	}
	rows := t.indexed().byName[sel]
	if len(rows) > 1 {
		rows = d.inside(t, rows)
	}
	switch len(rows) {
	case 1:
		return rows[0], nil
	case 0:
		return 0, fmt.Errorf("no row of %s has key or name %q", t.category, sel)
	}
	keys := make([]string, len(rows))
	for i, r := range rows {
		if t.key < 0 {
			keys[i] = t.selectorSpelling(r)
		} else {
			keys[i] = t.cell(r, t.key)
		}
	}
	if t.key < 0 {
		return 0, fmt.Errorf("%q names %d rows of %s; select it inside its %s, one of %v", sel, len(rows), t.category, t.parentT.category, keys)
	}
	inside := ""
	if t.parentT != nil {
		inside = fmt.Sprintf(", or select it inside its %s", t.parentT.category)
	}
	return 0, fmt.Errorf("%q names %d rows of %s; select one by key, one of %v%s", sel, len(rows), t.category, keys, inside)
}

// selectorSpelling is how a path writes a selected row, for messages: by key, or
// by name inside its parent's row.
func (t *table) selectorSpelling(r int) string {
	switch {
	case t.key >= 0:
		return t.category + "[" + t.cell(r, t.key) + "]"
	case t.name >= 0:
		return t.parentT.selectorSpelling(t.parentRow(r)) + "." + t.category + "[" + t.cell(r, t.name) + "]"
	}
	return fmt.Sprintf("%s line %d", t.file, r+2)
}
