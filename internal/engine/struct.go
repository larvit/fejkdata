package engine

import (
	"fmt"
	"reflect"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

// LoadStruct compiles the fake tags of the struct type t as one record, loads it against root,
// and proves each column holds only what its field's Go type can. A failed lazy load matches
// ErrLoad.
func LoadStruct(root *tree.Folder, t reflect.Type, label string, tags map[string]jsonvalue.Value) (Record, error) {
	l := tree.NewLoadFinds()
	n, err := load.CompileStruct(tags, l)
	if err != nil {
		return Record{}, fmt.Errorf("%s: %w", label, err)
	}
	if err := loadInline(root, n, l, label, &fence.GoStruct{Label: label, Type: t, Record: n}); err != nil {
		return Record{}, noDataNote(root, err)
	}
	record, err := recordOf(n)
	if err != nil {
		return Record{}, fmt.Errorf("%s: %w", label, err)
	}
	return Record{template: record}, nil
}

// FieldDataType is what a field of the Go kind k holds, and false where no fake tag fills that
// kind.
func FieldDataType(k reflect.Kind) (datatype.DataType, bool) {
	kind, holds := fence.FieldKinds[k]
	return kind.Datatype, holds
}
