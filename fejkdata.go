// Package fejkdata generates fake data from recursive JSON templates.
//
// Data lives in JSON, not in Go. A generator starts from the shipped data set and
// layers any directories you add; folders and files become a dot-path namespace,
// then generate values by path:
//
//	f, _ := fejkdata.New(fejkdata.WithSeed(42))
//	f.Fake("sv_SE.address")          // "Storgatan 12\n234 56 Göteborg"
//	f.Fake("sv_SE.address.locality") // "Göteborg"
//
// Several sources merge in order, the last winning a name clash, so custom data
// layers over the built-ins. The JSON template format is documented in the README.
package fejkdata

import (
	crand "crypto/rand"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"reflect"
	"sort"
	"sync"
)

//go:embed data
var shippedFS embed.FS

// MaxRepeat caps a repeat, and the renders nested repeats multiply to along any
// path; the CLI's --repeat shares it.
const MaxRepeat = 1 << 20

var _ [^uint(0)>>63 - 1]struct{} // 64-bit only, per the README's Decisions

// ErrNoData is returned by New when no source is loaded at all.
var ErrNoData = errors.New("no data: WithoutShippedData needs at least one WithDataPath or WithDataFS")

// Generator generates fake data from a loaded namespace tree. Create one with [New].
// It is safe for concurrent use; a seeded sequence is reproducible only when drawn
// from one goroutine. The compiled tree is immutable after [New], and Fake,
// NewTemplate and List read it concurrently without a lock.
type Generator struct {
	mu         sync.Mutex
	rand       *session
	categories map[string]node
	root       *folder // the categories as the node a path walks from
	set        drawSet // one Fake's draws, owned here so a walk pinning rows keeps them off the heap
	records    map[node]recordShape
	structs    map[reflect.Type]structResult
}

// session is one generator's mutable render state: the seeded rng plus the {seq()}
// counters.
type session struct {
	*rand.Rand
	counters map[string]uint64
}

func (s *session) next(key string) uint64 {
	s.counters[key]++
	return s.counters[key]
}

type config struct {
	seed    uint64
	seeded  bool
	shipped bool
	sources []dataSource
}

// Option configures a [Generator].
type Option func(*config)

// WithSeed makes output reproducible: two generators with the same seed and data
// emit identical sequences.
func WithSeed(seed uint64) Option {
	return func(c *config) { c.seed, c.seeded = seed, true }
}

// WithDataPath layers a data directory over what is loaded before it. Repeat it to
// layer several; the last wins a name clash.
func WithDataPath(dir string) Option {
	return func(c *config) {
		c.sources = append(c.sources, dataSource{fsys: os.DirFS(dir), label: dir, onDisk: true, path: dir})
	}
}

// WithDataFS layers a data tree held in an [fs.FS], such as an embed.FS of your own.
func WithDataFS(fsys fs.FS) Option {
	return func(c *config) { c.sources = append(c.sources, dataSource{fsys: fsys}) }
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
// or no data at all.
func New(opts ...Option) (*Generator, error) {
	c := config{shipped: true}
	for _, opt := range opts {
		opt(&c)
	}
	var sources []dataSource
	if c.shipped {
		sources = append(sources, dataSource{fsys: shippedFS, root: "data"})
	}
	sources = append(sources, c.sources...)
	if len(sources) == 0 {
		return nil, fmt.Errorf("fejkdata: %w", ErrNoData)
	}
	cats, err := loadData(sources)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	rng, err := newRand(c.seed, c.seeded)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &Generator{rand: rng, categories: cats, root: &folder{children: cats}}, nil
}

// List returns the sorted dotted paths Fake can render: every category, the dotted
// fields within a template, and folder segments. A choice consumes no segment, so a
// path continues through one only where every variant carries it, which is the
// rule Fake applies too: List is the set of paths Fake accepts.
func (f *Generator) List() []string {
	var out []string
	for _, name := range sortedNames(f.categories) {
		for _, p := range paths(f.categories[name]) {
			out = append(out, join(name, p))
		}
	}
	sort.Strings(out)
	return out
}

// paths lists the dot paths addressable from n, relative to it, where "" is n
// itself. A folder has no value of its own, so it contributes only its children's.
func paths(n node) []string {
	switch n := n.(type) {
	case *folder:
		var out []string
		for _, name := range sortedNames(n.children) {
			for _, p := range paths(n.children[name]) {
				out = append(out, join(name, p))
			}
		}
		return out
	case *template:
		out := []string{""}
		for _, name := range sortedNames(n.fields) {
			if isRef(name) { // a binding, not a path segment
				continue
			}
			for _, p := range paths(n.fields[name]) {
				out = append(out, join(name, p))
			}
		}
		return out
	case *null:
		return []string{""}
	case *table:
		return tablePaths(n)
	case *column:
		return []string{""}
	case *choice:
		out := []string{""}
		for p := range n.shared {
			out = append(out, p)
		}
		return out
	}
	return nil
}

// tablePaths is a table's columns, then each table linked to it under its name: the
// direct descents, a step at a time.
func tablePaths(t *table) []string {
	out := append([]string{""}, t.columns...)
	sort.Strings(out[1:])
	children := make([]string, 0, len(t.children))
	for name := range t.children {
		children = append(children, name)
	}
	sort.Strings(children)
	for _, name := range children {
		for _, p := range paths(t.children[name]) {
			out = append(out, join(name, p))
		}
	}
	return out
}

// sharedPaths is the sub-paths every item carries — the only ones a path may step
// through a choice to reach. It intersects, bailing as soon as the set
// is empty, which is immediate for a choice of plain strings.
func sharedPaths(items []node) map[string]bool {
	shared := subPaths(items[0])
	for _, it := range items[1:] {
		if len(shared) == 0 {
			return nil
		}
		next := subPaths(it)
		for p := range shared {
			if !next[p] {
				delete(shared, p)
			}
		}
	}
	return shared
}

// subPaths is paths(n) as a set, without the empty path that means n itself.
func subPaths(n node) map[string]bool {
	out := map[string]bool{}
	for _, p := range paths(n) {
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

// randomBytes seeds an unseeded generator.
var randomBytes = crand.Read

func newRand(seed uint64, seeded bool) (*session, error) {
	var r *rand.Rand
	if seeded {
		r = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	} else {
		var b [16]byte
		if _, err := randomBytes(b[:]); err != nil {
			return nil, fmt.Errorf("seeding from crypto/rand: %w", err)
		}
		r = rand.New(rand.NewPCG(binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:])))
	}
	return &session{Rand: r, counters: map[string]uint64{}}, nil
}
