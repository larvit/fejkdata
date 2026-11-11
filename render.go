package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine"
)

// Fake generates a value for a dot path. Each segment descends one level: folder
// names and the category (JSON file) come first, then named fields within it,
// e.g. "sv_SE.address" or "sv_SE.address.street". Choices along the way are
// drawn at random. A path naming a folder (no value of its own) is an error.
func (f *Generator) Fake(path string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, err := engine.Fake(f.drawState, &f.root, path)
	if err != nil {
		return "", fmt.Errorf("fejkdata: %w", err)
	}
	return v, nil
}
