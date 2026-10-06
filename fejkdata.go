package fejkdata

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"sync"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
)

// MaxRepeat caps a repeat, and caps the renders nested repeats multiply to along
// any one path; the CLI's --repeat shares it.
const MaxRepeat = 1 << 20

var _ [^uint(0)>>63 - 1]struct{} // docs/decisions.md#64-bit-targets-only

// ErrNoData is returned by New when no source is loaded at all.
var ErrNoData = errors.New("no data: WithoutShippedData needs at least one WithDataPath or WithDataFS")

// Generator generates fake data from a namespace tree. Create one with [New].
// It is safe for concurrent use; a seeded sequence is reproducible only when drawn
// from one goroutine.
type Generator struct {
	// mu guards drawState, structs and root, which gains a shipped category on the
	// first call reaching it.
	mu        sync.Mutex
	drawState *drawstate.State
	root      folder // the categories as the node a path walks from, owned here so a walk allocates none
	structs   map[reflect.Type]structResult
}

type config struct {
	seed    uint64
	seeded  bool
	shipped bool
	sources []datafiles.Source
}

// Option configures a [Generator].
type Option func(*config)

// WithSeed makes output reproducible: two generators with the same seed, version and
// data emit identical sequences.
func WithSeed(seed uint64) Option {
	return func(c *config) { c.seed, c.seeded = seed, true }
}

// WithDataPath layers a data directory over what is loaded before it. Repeat it to
// layer several; the last wins a name clash.
func WithDataPath(dir string) Option {
	return func(c *config) {
		c.sources = append(c.sources, datafiles.Dir(dir))
	}
}

// WithDataFS layers a data tree held in an [fs.FS], such as an embed.FS of your own.
func WithDataFS(fsys fs.FS) Option {
	return func(c *config) { c.sources = append(c.sources, datafiles.FS(fsys, "")) }
}

// WithoutShippedData leaves the shipped data set out, so only the sources given
// with [WithDataPath] and [WithDataFS] load.
func WithoutShippedData() Option {
	return func(c *config) { c.shipped = false }
}

// New builds a generator from the shipped data set and the options' sources, merged
// in order with the last winning a name clash. Each JSON file becomes a category
// named after the file (address.json -> "address") and each subdirectory a
// namespace segment. It errors on a missing directory, invalid JSON, invalid data,
// or no data at all. With only the shipped data, a category loads on the first call
// reaching it; with WithDataPath or WithDataFS, every category loads here.
func New(opts ...Option) (*Generator, error) {
	c := config{shipped: true}
	for _, opt := range opts {
		opt(&c)
	}
	root, err := c.load()
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	seed, err := c.drawSeed()
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &Generator{drawState: drawstate.New(seed), root: root}, nil
}

// docs/decisions.md#with-only-the-shipped-set-a-category-loads-on-the-first-call-reaching-it-beside-a---data-path-every-category-loads-in-new
func (c config) load() (folder, error) {
	if c.shipped && len(c.sources) == 0 {
		return unloadedTree(), nil
	}
	var sources []datafiles.Source
	if c.shipped {
		sources = append(sources, shippedSource)
	}
	sources = append(sources, c.sources...)
	if len(sources) == 0 {
		return folder{}, ErrNoData
	}
	cats, err := loadData(sources)
	return folder{children: cats}, err
}

// List returns the sorted dotted paths Fake renders: each category and every field,
// column and linked table below it, a direct descent at a time, listing a level
// carrying a repeat but nothing below it. A choice consumes no segment, so a path
// continues through one only where every variant carries it.
func (f *Generator) List() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := paths(&f.root, false)
	sort.Strings(out)
	return out
}

// paths lists the dot paths addressable from n, relative to it, where "" is n
// itself. A folder has no value of its own, so it contributes only its children's.
// With intoRepeats, it also lists the paths below a level carrying a repeat, so a path
// there meets the repeat's refusal.
func paths(n node, intoRepeats bool) []string {
	switch n := n.(type) {
	case *folder:
		var out []string
		for _, name := range sortedNames(n.children) {
			out = appendUnder(out, name, paths(n.children[name], intoRepeats))
		}
		for _, name := range sortedNames(n.unloaded) {
			out = appendUnder(out, name, n.unloaded[name].paths)
		}
		return out
	case *template:
		out := []string{""}
		if n.repeat > 1 && !intoRepeats {
			return out
		}
		for _, name := range sortedNames(n.fields) {
			out = appendUnder(out, name, paths(n.fields[name], intoRepeats))
		}
		return out
	case *table:
		return tablePaths(n, intoRepeats)
	case *tableColumn, *tableRow, *nullItem:
		return []string{""}
	case *choice:
		shared := n.shared
		if !intoRepeats {
			shared = sharedPaths(n.items, false)
		}
		return append([]string{""}, sortedNames(shared)...)
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
func tablePaths(t *table, intoRepeats bool) []string {
	out := append([]string{""}, t.rows.Header()...)
	sort.Strings(out[1:])
	for _, c := range t.rows.Children() {
		out = appendUnder(out, c.Segment(), paths(c.Owner(), intoRepeats))
	}
	return out
}

// sharedPaths is the sub-paths every item carries — the only ones a path may step
// through a choice to reach. It intersects, bailing as soon as the set
// is empty, which is immediate for a choice of plain strings.
func sharedPaths(items []node, intoRepeats bool) map[string]bool {
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

// subPaths is paths(n, intoRepeats) as a set, without the empty path that means n itself.
func subPaths(n node, intoRepeats bool) map[string]bool {
	out := map[string]bool{}
	for _, p := range paths(n, intoRepeats) {
		if p != "" {
			out[p] = true
		}
	}
	return out
}

func join(prefix, name string) string {
	switch {
	case prefix == "":
		return name
	case name == "":
		return prefix
	}
	return prefix + "." + name
}

var randomBytes = crand.Read

// drawSeed is the seed WithSeed gave, or one read from the system's entropy.
func (c config) drawSeed() (uint64, error) {
	if c.seeded {
		return c.seed, nil
	}
	var b [8]byte
	if _, err := randomBytes(b[:]); err != nil {
		return 0, fmt.Errorf("seeding from crypto/rand: %w", err)
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}
