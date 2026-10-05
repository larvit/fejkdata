// Package datatype names what a record column holds.
package datatype

import "fmt"

// DataType is what a record column holds, which decides how a record writes its value.
type DataType int

// The datatypes a column declares with "datatype"; a column without one is a string.
const (
	String DataType = iota
	Integer
	Number
	Boolean
)

// Count is the number of datatypes, for an array indexed by one.
const Count = Boolean + 1

var (
	names = [Count]string{"string", "integer", "number", "boolean"}
	nouns = [Count]string{"text", "an integer", "a number", "a boolean"}
)

// String is the datatype as data spells it.
func (d DataType) String() string {
	if d < 0 || d >= Count {
		return fmt.Sprintf("DataType(%d)", int(d))
	}
	return names[d]
}

// Noun is the datatype as an error names what a value is.
func Noun(d DataType) string { return nouns[d] }
