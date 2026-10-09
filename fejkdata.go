package fejkdata

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"sync"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/drawstate"
)

// MaxRepeat caps a repeat, and caps the renders nested repeats multiply to along
// any one path; the CLI's --repeat shares it.
const MaxRepeat = 1 << 20

// Fails to compile where uint is narrower than 64 bits: docs/decisions.md#64-bit-targets-only
var _ [^uint(0)>>63 - 1]struct{}

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
// namespace segment. It errors on a missing directory, invalid JSON or invalid data; a
// generator with no data at all renders templates that read none. With only the shipped data, a category loads on the first call
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

// load builds the tree New starts from: with only the shipped set, every category unloaded; else
// every source loaded and merged. docs/decisions.md#with-only-the-shipped-set-a-category-loads-on-the-first-call-reaching-it-beside-a---data-path-every-category-loads-in-new
func (c config) load() (folder, error) {
	if c.shipped && len(c.sources) == 0 {
		return unloadedTree(), nil
	}
	var sources []datafiles.Source
	if c.shipped {
		sources = append(sources, shippedSource)
	}
	sources = append(sources, c.sources...)
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
