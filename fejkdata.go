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
	// mu guards drawState, structs and root, which gains an indexed category on the
	// first call reaching it.
	mu        sync.Mutex
	drawState *drawstate.State
	root      folder // the categories as the node a path walks from, owned here so a walk allocates none
	structs   map[reflect.Type]structResult
}

type config struct {
	seed    uint64
	seeded  bool
	sources []datafiles.Source
	err     error
	// given is whether an option named the data to load, though it may name none.
	given bool
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
		c.sources, c.given = append(c.sources, datafiles.Dir(dir)), true
	}
}

// WithDataFS layers modules, each a data tree held in an [fs.FS], over what is loaded
// before them, in order. A module may be a data package's FS, such as one of
// github.com/larvit/fejkdata/data's Modules, or an embed.FS of your own. WithDataFS()
// loads no data.
func WithDataFS(modules ...fs.FS) Option {
	return func(c *config) {
		for i, fsys := range modules {
			if fsys == nil {
				c.err = errors.Join(c.err, fmt.Errorf("fejkdata: WithDataFS module %d is nil", i+1))
				continue
			}
			c.sources = append(c.sources, datafiles.FS(fsys))
		}
		c.given = true
	}
}

var errNoData = errors.New("fejkdata: New was given no data to load; pass WithDataFS(data.Modules()...) for every module fejkdata ships (import github.com/larvit/fejkdata/data); WithDataFS or WithDataPath for your own; or WithDataFS() for none")

// ErrLoad marks data that fails to load: in New, or on the first call reaching a category
// of an indexed source.
var ErrLoad = errors.New("data fails to load")

// loadError is a load's error, unchanged in its text, matching ErrLoad.
type loadError struct{ error }

func (e loadError) Unwrap() error      { return e.error }
func (loadError) Is(target error) bool { return target == ErrLoad }

// New builds a generator from the options' sources, merged in order with the last
// winning a name clash; it fails where no [WithDataFS] or [WithDataPath] names the data
// to load. Each JSON file becomes a category named after the file (address.json ->
// "address") and each subdirectory a namespace segment. It errors on a missing
// directory, invalid JSON or invalid data; a generator with no data at all renders
// templates that read none. A source whose .fejkdata.json carries an index, as each
// shipped module's does, loads each category on the first call reaching it. Data failing
// to load, in New or on that first call, gives an error matching [ErrLoad].
func New(opts ...Option) (*Generator, error) {
	var c config
	for _, opt := range opts {
		opt(&c)
	}
	if !c.given {
		return nil, errNoData
	}
	if c.err != nil {
		return nil, c.err
	}
	root, err := loadSources(c.sources)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", loadError{err})
	}
	seed, err := c.drawSeed()
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &Generator{drawState: drawstate.New(seed), root: root}, nil
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
