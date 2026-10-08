// Package rows holds a table's rows: the TSV they are read from, the options proved over
// them, the links between tables, and the rows a draw takes or a selector names.
package rows

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
)

// Table is the rows of one table category, and O is what owns them.
type Table[O any] struct {
	owner       O
	segment     string // path's last segment
	path        string // the category's path from the data root, which a selector is written at
	file        string
	options     Options
	header      []string
	col         map[string]int
	cells       []string // rows × columns, flat
	keyIndex    int      // keyIndex through parentIndex are column indexes, -1 where the option is absent
	nameIndex   int
	weightIndex int
	parentIndex int
	weights     []float64 // nil when uniform
	cum         []float64 // cumulative weights, nil when uniform
	byKey       map[string]int
	parent      *Table[O]
	children    map[string]*Table[O]
	once        sync.Once
	lookup      rowLookup
}

// Options names the column each option reads, "" where the option is absent.
type Options struct {
	Key, Name, Parent, Weight string
}

// rowLookup is what selection and linked draws look up, built on the first draw
// that needs it.
type rowLookup struct {
	byName       map[string][]int
	rowsByParent map[string][]int
	childCum     map[string][]float64 // cumulative weights per parent key, nil when uniform
}

// Parse reads the TSV data, the rows of the table at path, and proves what the options
// claim of them.
func Parse[O any](owner O, segment, path, file, data string, o Options) (*Table[O], error) {
	t := &Table[O]{owner: owner, segment: segment, path: path, file: file, options: o, keyIndex: -1, nameIndex: -1, weightIndex: -1, parentIndex: -1}
	if err := t.parseRows(data); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if err := t.bindOptions(); err != nil {
		return nil, err
	}
	if err := t.checkSelectorCells(); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return t, nil
}

func (t *Table[O]) Owner() O          { return t.owner }
func (t *Table[O]) Segment() string   { return t.segment }
func (t *Table[O]) File() string      { return t.file }
func (t *Table[O]) Options() Options  { return t.options }
func (t *Table[O]) Header() []string  { return t.header }
func (t *Table[O]) Parent() *Table[O] { return t.parent }

func (t *Table[O]) Column(name string) (int, bool) {
	i, ok := t.col[name]
	return i, ok
}

func (t *Table[O]) Len() int { return len(t.cells) / len(t.header) }

func (t *Table[O]) Cell(row, col int) string { return t.cells[row*len(t.header)+col] }

// parseRows reads the TSV: the header line names the columns, each following line
// is a row of as many cells. Cells are substrings of data, so the file is held once.
func (t *Table[O]) parseRows(data string) error {
	data = strings.TrimSuffix(strings.TrimPrefix(data, "\xEF\xBB\xBF"), "\n")
	header, rest, _ := strings.Cut(data, "\n")
	header = strings.TrimSuffix(header, "\r")
	if header == "" {
		return fmt.Errorf("has no header line naming its columns")
	}
	t.header = strings.Split(header, "\t")
	t.col = make(map[string]int, len(t.header))
	for i, name := range t.header {
		if err := grammar.CheckIdentifier(name); err != nil {
			return fmt.Errorf("column %w", err)
		}
		if _, dup := t.col[name]; dup {
			return fmt.Errorf("column %q is named twice", name)
		}
		t.col[name] = i
	}
	if header == data {
		return fmt.Errorf("has no rows below its header; a table holds at least one row")
	}
	m := len(t.header)
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
	return nil
}

// appendRow splits one line into as many cells as the header has columns.
func (t *Table[O]) appendRow(row string, line int) error {
	if row == "" {
		return fmt.Errorf("line %d is empty; every line below the header is a row", line)
	}
	m := len(t.header)
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

// bindOptions resolves each option to its column, and proves what the options claim of the cells: a
// key is unique, and a weight is a number of 0 or more.
func (t *Table[O]) bindOptions() error {
	o := t.options
	for _, opt := range []struct {
		name, value string
		into        *int
	}{{"key", o.Key, &t.keyIndex}, {"name", o.Name, &t.nameIndex}, {"parent", o.Parent, &t.parentIndex}, {"weight", o.Weight, &t.weightIndex}} {
		if opt.value == "" {
			continue
		}
		i, ok := t.col[opt.value]
		if !ok {
			return fmt.Errorf("%s names no column %q of %s; the columns are %v", opt.name, opt.value, t.file, t.header)
		}
		*opt.into = i
	}
	if t.nameIndex >= 0 && t.keyIndex < 0 && t.parentIndex < 0 {
		return fmt.Errorf("name selects a row as a key does, and lists the rows it matches by their keys, so it needs a key column, or a parent inside which each name is one row; add key")
	}
	if err := t.proveKeys(); err != nil {
		return err
	}
	if err := t.proveNamesInsideParent(); err != nil {
		return err
	}
	return t.sumWeights()
}

// proveNamesInsideParent proves a name without a key names one row inside its parent.
func (t *Table[O]) proveNamesInsideParent() error {
	if t.keyIndex >= 0 || t.nameIndex < 0 {
		return nil
	}
	inside := make(map[string]int, t.Len())
	for r := 0; r < t.Len(); r++ {
		k := t.Cell(r, t.parentIndex) + "\t" + t.Cell(r, t.nameIndex)
		if first, dup := inside[k]; dup {
			return fmt.Errorf("%s line %d: name %q repeats line %d inside %s %q; a name selects one row inside its parent; drop one, or add a key column", t.file, r+2, t.Cell(r, t.nameIndex), first+2, t.header[t.parentIndex], t.Cell(r, t.parentIndex))
		}
		inside[k] = r
	}
	return nil
}

// proveKeys proves every key names one row, and keeps the map a link is proved by.
func (t *Table[O]) proveKeys() error {
	if t.keyIndex < 0 {
		return nil
	}
	t.byKey = make(map[string]int, t.Len())
	for r := 0; r < t.Len(); r++ {
		k := t.Cell(r, t.keyIndex)
		if k == "" {
			return fmt.Errorf("%s line %d: the key is empty", t.file, r+2)
		}
		if first, dup := t.byKey[k]; dup {
			return fmt.Errorf("%s line %d: key %q repeats line %d; a key selects one row", t.file, r+2, k, first+2)
		}
		t.byKey[k] = r
	}
	for r := 0; r < t.Len() && t.nameIndex >= 0; r++ {
		if n := t.Cell(r, t.nameIndex); t.byKey[n] != r {
			if other, isKey := t.byKey[n]; isKey {
				return fmt.Errorf("%s line %d: name %q is the key of line %d, which a selector reads first, so the name could never select this row", t.file, r+2, n, other+2)
			}
		}
	}
	return nil
}

// sumWeights proves every weight is a number of 0 or more, some above 0, and keeps the weights
// and their running sum.
func (t *Table[O]) sumWeights() error {
	if t.weightIndex < 0 {
		return nil
	}
	t.weights, t.cum = make([]float64, t.Len()), make([]float64, t.Len())
	total := 0.0
	for r := range t.cum {
		cell := t.Cell(r, t.weightIndex)
		w, err := strconv.ParseFloat(cell, 64)
		if err != nil || math.IsInf(w, 0) || math.IsNaN(w) || w < 0 {
			return fmt.Errorf("%s line %d: weight %q is not a number of 0 or more", t.file, r+2, cell)
		}
		if grammar.Underflows(cell, w) {
			return fmt.Errorf("%s line %d: weight %s is too close to 0 to tell from it", t.file, r+2, cell)
		}
		total += w
		t.weights[r], t.cum[r] = w, total
	}
	switch {
	case math.IsInf(total, 0):
		return fmt.Errorf("%s: the weights sum past the largest number", t.file)
	case total == 0:
		return fmt.Errorf("%s: every weight is 0, so no row is drawn", t.file)
	}
	return t.proveWeightUnderEachParent()
}

// proveWeightUnderEachParent proves each parent key's rows weigh more than 0 together, so a draw
// inside that parent's row has a row to draw.
func (t *Table[O]) proveWeightUnderEachParent() error {
	if t.parentIndex < 0 {
		return nil
	}
	sums, keys := map[string]float64{}, []string{}
	for r := range t.weights {
		k := t.Cell(r, t.parentIndex)
		if _, seen := sums[k]; !seen {
			keys = append(keys, k)
		}
		sums[k] += t.weights[r]
	}
	for _, k := range keys {
		if sums[k] == 0 {
			return fmt.Errorf("%s: every row under %s %q weighs 0, so none is drawn there", t.file, t.header[t.parentIndex], k)
		}
	}
	return nil
}

// checkSelectorCells refuses a key or name cell a selector cannot spell.
func (t *Table[O]) checkSelectorCells() error {
	for r := 0; r < t.Len(); r++ {
		for _, col := range []int{t.keyIndex, t.nameIndex} {
			if col < 0 {
				continue
			}
			if c := grammar.UnspellableInSelector(t.Cell(r, col)); c != "" {
				return fmt.Errorf("line %d: %s %q contains %q, which a selector cannot spell", r+2, t.header[col], t.Cell(r, col), c)
			}
		}
	}
	return nil
}

// Link links t to p, the table its link column names, after proving p has a key, every
// link cell is one, and every row of p is linked to. sibling returns the table beside t
// named name, or nil. Link is the only writer of a table's parent and children.
// docs/decisions.md#a-parent-row-with-no-child-row-is-a-load-error
func (t *Table[O]) Link(p *Table[O], sibling func(name string) *Table[O]) error {
	name := t.header[t.parentIndex]
	if p.keyIndex < 0 {
		return fmt.Errorf("parent %q has no key column to link to", name)
	}
	if err := t.checkAncestors(p, sibling); err != nil {
		return err
	}
	linked := make(map[string]bool, p.Len())
	for r := 0; r < t.Len(); r++ {
		k := t.Cell(r, t.parentIndex)
		if _, ok := p.byKey[k]; !ok {
			return fmt.Errorf("%s line %d: %s %q is no key of %s", t.file, r+2, name, k, p.file)
		}
		linked[k] = true
	}
	for r := 0; r < p.Len(); r++ {
		if k := p.Cell(r, p.keyIndex); !linked[k] {
			return fmt.Errorf("%s links no row to %s %q; every %s row needs one, or drop line %d of %s", t.file, name, k, name, r+2, p.file)
		}
	}
	t.parent = p
	if p.children == nil {
		p.children = map[string]*Table[O]{}
	}
	p.children[t.segment] = t
	return nil
}

// checkAncestors refuses a cycle in the chain of parents starting at p, t's parent, and
// refuses an ancestor that has a column named like t.
func (t *Table[O]) checkAncestors(p *Table[O], sibling func(name string) *Table[O]) error {
	var ancestors []*Table[O]
	for q, seen := p, map[*Table[O]]bool{t: true}; q != nil; q = sibling(q.header[q.parentIndex]) {
		if seen[q] {
			return fmt.Errorf("parent cycle: %s reaches itself through its parents", q.segment)
		}
		seen[q] = true
		ancestors = append(ancestors, q)
		if q.parentIndex < 0 {
			break
		}
	}
	for _, q := range ancestors {
		if _, clash := q.col[t.segment]; clash {
			return fmt.Errorf("%q is named like a column of %q, its ancestor, so %s.%s could read either; rename one", t.segment, q.segment, q.segment, t.segment)
		}
	}
	return nil
}

// Children is the tables linked to t, sorted by segment.
func (t *Table[O]) Children() []*Table[O] {
	out := make([]*Table[O], 0, len(t.children))
	for _, c := range t.children {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].segment < out[j].segment })
	return out
}

// Descendant is the table named name among those linked to t, at any depth.
func (t *Table[O]) Descendant(name string) *Table[O] {
	if c, ok := t.children[name]; ok {
		return c
	}
	for _, c := range t.children {
		if d := c.Descendant(name); d != nil {
			return d
		}
	}
	return nil
}

// ProveStepUp proves a ".." after a row of t can step up: t has a parent, rest has no empty
// segment and names it, and no selector follows, since a row selected below the parent row
// could lie outside it.
func (t *Table[O]) ProveStepUp(rest []string) error {
	if t.parent == nil {
		return fmt.Errorf(`%s has no parent table for ".." to step up to`, t.segment)
	}
	if err := grammar.CheckSegments(rest); err != nil {
		return err
	}
	switch {
	case len(rest) == 0 || rest[0] != t.parent.segment:
		return fmt.Errorf(`".." steps up from %s to its parent table, so name that next: %s`, t.segment, t.climbTo(rest))
	case grammar.HasSelector(rest):
		return fmt.Errorf(`no selector after "..": a row selected there could lie outside the row stepped up to; select from the table's path from the data root: %s`, grammar.JoinSegments(append([]string{t.parent.path}, rest[1:]...)))
	}
	return nil
}

// climbTo spells the ".." steps from t up to the ancestor rest names first, or the one step to
// t's parent where no ancestor is named.
func (t *Table[O]) climbTo(rest []string) string {
	var b strings.Builder
	for a := t.parent; a != nil; a = a.parent {
		b.WriteString(".." + a.segment)
		if len(rest) > 0 && a.segment == rest[0] {
			return b.String()
		}
	}
	return ".." + t.parent.segment
}

func (t *Table[O]) builtLookup() *rowLookup {
	t.once.Do(func() {
		if t.nameIndex >= 0 {
			t.lookup.byName = make(map[string][]int, t.Len())
			for r := 0; r < t.Len(); r++ {
				n := t.Cell(r, t.nameIndex)
				t.lookup.byName[n] = append(t.lookup.byName[n], r)
			}
		}
		if t.parentIndex >= 0 {
			t.lookup.rowsByParent = map[string][]int{}
			for r := 0; r < t.Len(); r++ {
				k := t.Cell(r, t.parentIndex)
				t.lookup.rowsByParent[k] = append(t.lookup.rowsByParent[k], r)
			}
			if t.weights != nil {
				t.lookup.childCum = make(map[string][]float64, len(t.lookup.rowsByParent))
				for k, rows := range t.lookup.rowsByParent {
					cum, total := make([]float64, len(rows)), 0.0
					for i, r := range rows {
						total += t.weights[r]
						cum[i] = total
					}
					t.lookup.childCum[k] = cum
				}
			}
		}
	})
	return &t.lookup
}

func (t *Table[O]) Draw(s *drawstate.State) int {
	if t.cum == nil {
		return s.IntN(t.Len())
	}
	return s.Weighted(t.cum)
}

// drawUnder draws a row among those linked to parent row pr.
func (t *Table[O]) drawUnder(s *drawstate.State, pr int) int {
	lookup := t.builtLookup()
	k := t.parent.Cell(pr, t.parent.keyIndex)
	rows := lookup.rowsByParent[k]
	if lookup.childCum == nil {
		return rows[s.IntN(len(rows))]
	}
	return rows[s.Weighted(lookup.childCum[k])]
}

// DrawIn returns the row of t that p pins. Where p pins none, it draws one inside the nearest
// ancestor p pins, drawing t's parent there first when that ancestor is higher up; with no
// ancestor pinned, it draws over the whole table. It pins the row drawn and its ancestors' rows in p.
func (t *Table[O]) DrawIn(s *drawstate.State, p *Pins[O]) int {
	if r, ok := p.Pinned(t); ok {
		return r
	}
	r := -1
	for a := t.parent; a != nil && r < 0; a = a.parent {
		if _, ok := p.Pinned(a); !ok {
			continue
		}
		if t.parent != a {
			t.parent.DrawIn(s, p)
		}
		pr, _ := p.Pinned(t.parent)
		r = t.drawUnder(s, pr)
	}
	if r < 0 {
		r = t.Draw(s)
	}
	p.pin(t, r)
	return r
}

// parentRow is the row of t's parent that row r links to.
func (t *Table[O]) parentRow(r int) int { return t.parent.byKey[t.Cell(r, t.parentIndex)] }

// under reports whether row r of t sits inside row pr of ancestor a.
func (t *Table[O]) under(r int, a *Table[O], pr int) bool {
	for c, row := t, r; c.parent != nil; c, row = c.parent, c.parentRow(row) {
		if c.parent == a {
			return c.parentRow(row) == pr
		}
	}
	return false
}

// find is the row a selector names: by key first, then by name, where a name
// naming several rows resolves inside the ancestors pinned.
func (t *Table[O]) find(sel string, pins *Pins[O]) (int, error) {
	if t.keyIndex < 0 && t.nameIndex < 0 {
		return 0, fmt.Errorf("%s has no key or name column to select a row by", t.path)
	}
	if r, ok := t.byKey[sel]; ok {
		return r, nil
	}
	rows := t.builtLookup().byName[sel]
	if len(rows) > 1 {
		rows = pins.inside(t, rows)
	}
	switch len(rows) {
	case 1:
		return rows[0], nil
	case 0:
		return 0, fmt.Errorf("no row of %s has key or name %q", t.path, sel)
	}
	return 0, t.ambiguous(sel, rows)
}

// ambiguous names each row a selector could have meant: a keyed table by its keys,
// and one told apart only by its parent by the path that selects it.
func (t *Table[O]) ambiguous(sel string, rows []int) error {
	spellings := make([]string, len(rows))
	for i, r := range rows {
		if t.keyIndex < 0 {
			spellings[i] = t.selectorSpelling(r)
		} else {
			spellings[i] = t.Cell(r, t.keyIndex)
		}
	}
	listed := strings.Join(spellings, ", ")
	if t.keyIndex < 0 {
		return fmt.Errorf("%q names %d rows of %s; select it inside its %s, one of %s", sel, len(rows), t.path, t.parent.path, listed)
	}
	inside := ""
	if t.parent != nil {
		inside = fmt.Sprintf(", or select it inside its %s", t.parent.path)
	}
	return fmt.Errorf("%q names %d rows of %s; select one by key, one of %s%s", sel, len(rows), t.path, listed, inside)
}

// selectorSpelling is how a path writes a selected row, for messages: by key, or
// by name inside its parent's row.
func (t *Table[O]) selectorSpelling(r int) string {
	switch {
	case t.keyIndex >= 0:
		return t.path + "[" + t.Cell(r, t.keyIndex) + "]"
	case t.nameIndex >= 0:
		return t.parent.selectorSpelling(t.parentRow(r)) + "." + t.segment + "[" + t.Cell(r, t.nameIndex) + "]"
	}
	return fmt.Sprintf("%s line %d", t.file, r+2)
}
