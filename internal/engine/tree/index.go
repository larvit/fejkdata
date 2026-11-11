package tree

import (
	"github.com/larvit/fejkdata/internal/datafiles"
)

// Indexed is a category an index names and no call has reached yet, and the source holding it.
type Indexed struct {
	entry datafiles.IndexEntry
	src   datafiles.Source
}

func NewIndexed(entry datafiles.IndexEntry, src datafiles.Source) *Indexed {
	return &Indexed{entry, src}
}

func (i Indexed) Entry() datafiles.IndexEntry { return i.entry }

func (i Indexed) Src() datafiles.Source { return i.src }
