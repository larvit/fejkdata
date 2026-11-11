package load

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

// datatypeOf reads a template's "datatype", nil where it carries none or sits outside a column.
func datatypeOf(m map[string]any, pos position) (*datatype.DataType, error) {
	v, ok := m["datatype"]
	if !ok {
		return nil, nil
	}
	name, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("datatype must be a string, not %s", jsonvalue.Kind(v))
	}
	for d := datatype.String; d < datatype.Count; d++ {
		if name != d.String() {
			continue
		}
		if !pos.column() {
			return nil, nil
		}
		return &d, nil
	}
	return nil, fmt.Errorf(`datatype takes "string", "integer", "number" or "boolean", got %q`, name)
}
