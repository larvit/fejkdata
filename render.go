package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/render"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// Fake generates a value for a dot path. Each segment descends one level: folder
// names and the category (JSON file) come first, then named fields within it,
// e.g. "sv_SE.address" or "sv_SE.address.street". Choices along the way are
// drawn at random. A path naming a folder (no value of its own) is an error.
func (f *Generator) Fake(path string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path, segments, err := f.loadCallerPath(path)
	if err != nil {
		return "", err
	}
	n, pins, err := drawCallerPath(f.drawState, &f.root, segments)
	if err != nil {
		return "", f.noDataNote(fmt.Errorf("fejkdata: %s: %w", path, err))
	}
	if _, ok := n.(*tree.Folder); ok {
		return "", fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	return render.Pinned(f.drawState, n, &pins), nil
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
