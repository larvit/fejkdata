package tree

import (
	"github.com/larvit/fejkdata/internal/datatype"
)

// RecordColumn is one column of a record template: its name and datatype, the field drawing it,
// and the name the template binds to that whole field, whose pick the column renders so the row
// agrees with the columns reading the name.
type RecordColumn struct {
	name       string
	datatype   datatype.DataType
	field      Node
	boundWhole *NameBinding
}

func NewRecordColumn(name string, datatype datatype.DataType, field Node, boundWhole *NameBinding) RecordColumn {
	return RecordColumn{name, datatype, field, boundWhole}
}

func (r RecordColumn) BoundWhole() *NameBinding { return r.boundWhole }

func (r RecordColumn) DataType() datatype.DataType { return r.datatype }

func (r RecordColumn) Field() Node { return r.field }

func (r RecordColumn) Name() string { return r.name }
