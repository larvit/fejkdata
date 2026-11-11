package engine

import (
	"errors"
	"fmt"
	"sort"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/render"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// MaxRepeat caps a repeat, and caps the renders nested repeats multiply to along any one path.
const MaxRepeat = load.MaxRepeat

// Fake draws the value a caller's path names, first loading any unloaded category it reaches.
// A failed load matches ErrLoad.
func Fake(s *drawstate.State, root *tree.Folder, path string) (string, error) {
	path, segments, err := callerPath(root, path)
	if err != nil {
		return "", err
	}
	n, pins, err := drawCallerPath(s, root, segments)
	if err != nil {
		return "", noDataNote(root, fmt.Errorf("%s: %w", path, err))
	}
	if _, ok := n.(*tree.Folder); ok {
		return "", fmt.Errorf("%s names a folder, not a value", path)
	}
	return render.Pinned(s, n, &pins), nil
}

// drawCallerPath walks a caller's path to the node it names, returning the rows its leaf renders in.
func drawCallerPath(s *drawstate.State, root tree.Node, segments []string) (tree.Node, tree.PinSet, error) {
	// docs/decisions.md#a-path-is-walked-once-without-drawing-before-it-is-walked-for-real
	var buf [16]tree.PathStep
	steps, err := load.CallerPathSteps(root, segments, buf[:0])
	if err != nil {
		return nil, tree.PinSet{}, err
	}
	var pins tree.PinSet
	n, leafPins := render.Walk(s, root, steps, &pins)
	return n, *leafPins, nil
}

// List is the sorted dotted paths Fake renders.
func List(root *tree.Folder) []string {
	out := load.Paths(root, false)
	sort.Strings(out)
	return out
}

// noDataNote adds to err, where a segment names nothing and root holds no data at all, that no
// data is loaded.
func noDataNote(root *tree.Folder, err error) error {
	var ne load.NoEntry
	if errors.As(err, &ne) && len(root.Children()) == 0 && len(root.Unloaded()) == 0 {
		return fmt.Errorf("%w; no data is loaded", err)
	}
	return err
}
