package load

import (
	"fmt"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// ChildNamed is the child tree.Child finds, else an error saying why seg names nothing under n.
func ChildNamed(n tree.Node, seg string) (tree.Node, error) {
	switch {
	case grammar.IsSelector(seg):
		return nil, fmt.Errorf("%s is not a table, so it has no row to select", grammar.SelectorOf(seg))
	case seg == "..":
		if c, ok := n.(*tree.TableColumn); ok {
			return nil, fmt.Errorf(`".." steps up from a table's row, and %q is a column; put the ".." right after the row`, c.Name())
		}
		return nil, fmt.Errorf(`".." steps up only from a table's row; name any other level by its path from the data root`)
	}
	if child, ok := tree.Child(n, seg); ok {
		return child, nil
	}
	switch n := n.(type) {
	case *tree.Folder:
		return nil, NoEntry{seg}
	case *tree.Template:
	case *tree.TableColumn:
		if n.Name() == n.Table().Rows().Options().Parent {
			return nil, fmt.Errorf("no field %q: %q is the link column, holding its parent's key; to read the row it links to, step up with ..%s.%s", seg, n.Name(), n.Name(), seg)
		}
		return nil, fmt.Errorf("no field %q: %q is a column, and a cell holds no fields", seg, n.Name())
	case *tree.NullItem:
	default:
		panic(invariant.Broken("ChildNamed has no case for node %T", n))
	}
	return nil, fmt.Errorf("no field %q", seg)
}

// compiledPath walks a path and proves it reaches a node whichever way the draws go: every
// variant of a choice carries the rest of it, and is walked, a selector names a row inside the
// rows selected before it, and no level read carries a repeat. Once run, it holds the steps a
// draw takes, and every leaf the path may render. headSpans is how many of tail's first segments
// head stands for. An error names a level as head, then tail's segments after those, up to that level.
type compiledPath struct {
	pins      tree.PinSet
	head      string
	headSpans int
	tail      []string
	steps     []tree.PathStep
	Leaves    []tree.Node
}

func (w *compiledPath) run(n tree.Node) (tree.Node, error) {
	return w.walk(n, w.tail)
}

func (w *compiledPath) walk(n tree.Node, tail []string) (tree.Node, error) {
	at := pathPos{n: n, tail: tail}
	for at.more() {
		if c, ok := at.n.(*tree.Choice); ok {
			return w.walkEvery(c, at.tail)
		}
		var err error
		if w.steps, err = compileStep(&at, w.tail, w.head, w.headSpans, w.steps, &w.pins); err != nil {
			return nil, err
		}
	}
	w.Leaves = append(w.Leaves, at.n)
	return at.n, nil
}

func repeatLevelError(level string) error {
	return fmt.Errorf("the level %q carries a repeat, so a path cannot read one draw of it; read %q whole", level, level)
}

// pathPos is where a walk stands: the node reached, the tail left to walk from it, and fromRow,
// whether the walk entered n, a table, from a row of a table linked to it.
type pathPos struct {
	n       tree.Node
	tail    []string
	fromRow bool
}

func (p pathPos) more() bool { return len(p.tail) > 0 || p.fromRow }

// compileStep takes the first step of at.tail from at.n, which must not be a choice, moves at past
// it, and appends the step to steps: a field, or the route through a table. whole is the full
// path; head and headSpans spell an error's level, as in compiledPath.
func compileStep(at *pathPos, whole []string, head string, headSpans int, steps []tree.PathStep, pins *tree.PinSet) ([]tree.PathStep, error) {
	i := len(whole) - len(at.tail)
	switch x := at.n.(type) {
	case *tree.Table:
		r, err := routeThrough(x, at.tail, at.fromRow)
		if err != nil {
			return steps, err
		}
		*at = r.then
		return routeSteps(steps, pins, x, r, i)
	case *tree.Template:
		if x.Repeat() > 1 && !grammar.IsSelector(at.tail[0]) {
			return steps, repeatLevelError(join(head, grammar.JoinSegments(whole[headSpans:i])))
		}
	}
	next, err := ChildNamed(at.n, at.tail[0])
	if err != nil {
		return steps, err
	}
	steps = append(steps, tree.PathStep{Kind: tree.StepField, At: i, Name: at.tail[0]})
	*at = pathPos{n: next, tail: at.tail[1:]}
	return steps, nil
}

// walkEvery walks every variant and keeps the first one's steps: every variant carries
// the rest of the path, so each compiles to the same steps past the choice.
func (w *compiledPath) walkEvery(c *tree.Choice, tail []string) (tree.Node, error) {
	if err := carriedByAll(c, tail); err != nil {
		return nil, err
	}
	var last tree.Node
	// before has no spare capacity, so each variant's append copies and leaves the first variant's
	// steps intact.
	before, first := w.steps[:len(w.steps):len(w.steps)], w.steps
	for i, item := range c.Items() {
		var err error
		w.steps = before
		if last, err = w.walk(item, tail); err != nil {
			return nil, err
		}
		if i == 0 {
			first = w.steps
		}
	}
	w.steps = first
	return last, nil
}

// CallerPathSteps proves a path reaches a node without drawing, appending to steps the steps a
// draw takes. Where compiledPath walks every variant of a choice, it walks the first,
// which carriedByAll lets stand for all.
func CallerPathSteps(n tree.Node, tail []string, steps []tree.PathStep) ([]tree.PathStep, error) {
	var pins tree.PinSet
	for at := (pathPos{n: n, tail: tail}); at.more(); {
		if c, ok := at.n.(*tree.Choice); ok {
			if err := carriedByAll(c, at.tail); err != nil {
				return nil, err
			}
			at.n = c.Items()[0]
			continue
		}
		var err error
		if steps, err = compileStep(&at, tail, "", 0, steps, &pins); err != nil {
			return nil, err
		}
	}
	return steps, nil
}

// Paths lists the dot paths addressable from n, relative to it, where "" is n
// itself. A folder has no value of its own, so it contributes only its children's.
// With intoRepeats, it also lists the paths below a level carrying a repeat, so a path
// there meets the repeat's refusal.
func Paths(n tree.Node, intoRepeats bool) []string {
	switch n := n.(type) {
	case *tree.Folder:
		var out []string
		for _, name := range tree.SortedNames(n.Children()) {
			out = appendUnder(out, name, Paths(n.Children()[name], intoRepeats))
		}
		for _, name := range tree.SortedNames(n.Unloaded()) {
			out = appendUnder(out, name, n.Unloaded()[name].Entry().Paths)
		}
		return out
	case *tree.Template:
		out := []string{""}
		if n.Repeat() > 1 && !intoRepeats {
			return out
		}
		for _, name := range tree.SortedNames(n.Fields()) {
			out = appendUnder(out, name, Paths(n.Fields()[name], intoRepeats))
		}
		return out
	case *tree.Table:
		return tablePaths(n, intoRepeats)
	case *tree.TableColumn, *tree.TableRow, *tree.NullItem:
		return []string{""}
	case *tree.Choice:
		shared := n.Shared()
		if !intoRepeats {
			shared = sharedPaths(n.Items(), false)
		}
		return append([]string{""}, tree.SortedNames(shared)...)
	default:
		panic(invariant.Broken("paths has no case for node %T", n))
	}
}

func appendUnder(out []string, name string, ps []string) []string {
	for _, p := range ps {
		out = append(out, join(name, p))
	}
	return out
}

// tablePaths is a table's columns, then each table linked to it under its name: the
// direct descents, a step at a time.
func tablePaths(t *tree.Table, intoRepeats bool) []string {
	out := append([]string{""}, t.Rows().Header()...)
	sort.Strings(out[1:])
	for _, c := range t.ChildTables() {
		out = appendUnder(out, c.Rows().Segment(), Paths(c, intoRepeats))
	}
	return out
}

// CategoryPath is the dot path of the category name in the folders dir.
func CategoryPath(dir []string, name string) string { return join(strings.Join(dir, "."), name) }

func join(prefix, name string) string {
	switch {
	case prefix == "":
		return name
	case name == "":
		return prefix
	}
	return prefix + "." + name
}

// sharedPaths is the sub-paths every item carries — the only ones a path may step
// through a choice to reach. It intersects, bailing as soon as the set
// is empty, which is immediate for a choice of plain strings.
func sharedPaths(items []tree.Node, intoRepeats bool) map[string]bool {
	shared := subPaths(items[0], intoRepeats)
	for _, it := range items[1:] {
		if len(shared) == 0 {
			return nil
		}
		next := subPaths(it, intoRepeats)
		for p := range shared {
			if !next[p] {
				delete(shared, p)
			}
		}
	}
	return shared
}

// subPaths is Paths(n, intoRepeats) as a set, without the empty path that means n itself.
func subPaths(n tree.Node, intoRepeats bool) map[string]bool {
	out := map[string]bool{}
	for _, p := range Paths(n, intoRepeats) {
		if p != "" {
			out[p] = true
		}
	}
	return out
}

// carriedByAll is the choice rule a path that must reach a node on every call obeys:
// the rest of the tail must be one every variant carries.
func carriedByAll(c *tree.Choice, rest []string) error {
	if want := strings.Join(rest, "."); !c.Shared()[want] {
		return unreachableInChoice(c, want)
	}
	return nil
}

// unreachableInChoice reports that a path cannot step through this choice, listing
// what every variant does carry.
func unreachableInChoice(c *tree.Choice, want string) error {
	if len(c.Shared()) == 0 {
		return fmt.Errorf("no variant of this %d-way choice carries %q", len(c.Items()), want)
	}
	offered := tree.SortedNames(c.Shared())
	return fmt.Errorf("not every variant of this %d-way choice carries %q; all carry %v", len(c.Items()), want, offered)
}

func provePath(n tree.Node, tail []string, head string, headSpans int) error {
	_, err := (&compiledPath{head: head, headSpans: headSpans, tail: tail}).run(n)
	return err
}

// CompilePath compiles a path provePath proved.
func CompilePath(n tree.Node, tail []string) compiledPath {
	w := compiledPath{tail: tail}
	if _, err := w.run(n); err != nil {
		panic(invariant.Broken("%s: %v; the path should have been proved before it was compiled", grammar.JoinSegments(tail), err))
	}
	return w
}
