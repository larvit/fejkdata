package fejkdata

import (
	"errors"
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/load"
)

// noDataNote adds to err, where a segment names nothing and f holds no data at all, that no
// data is loaded.
func (f *Generator) noDataNote(err error) error {
	var ne load.NoEntry
	if errors.As(err, &ne) && len(f.root.Children()) == 0 && len(f.root.Unloaded()) == 0 {
		return fmt.Errorf("%w; no data is loaded", err)
	}
	return err
}
