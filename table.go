package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

func linkTables(sites []categorySite) error {
	for _, s := range sites {
		if t, isTable := s.n.(*tree.Table); isTable {
			if err := load.LinkTable(t, s.in.Children()); err != nil {
				return fmt.Errorf("%s: %w", s.path, err)
			}
		}
	}
	return nil
}
