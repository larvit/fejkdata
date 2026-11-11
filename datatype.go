package fejkdata

import "github.com/larvit/fejkdata/internal/datatype"

// DataType is what a record column holds, which decides how a record writes its value. It
// prints as data spells it: string, integer, number or boolean.
type DataType = datatype.DataType

// The datatypes a column declares with "datatype"; a column without one is a string.
const (
	DataTypeString  DataType = datatype.String
	DataTypeInteger DataType = datatype.Integer
	DataTypeNumber  DataType = datatype.Number
	DataTypeBoolean DataType = datatype.Boolean
)
