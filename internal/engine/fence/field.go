package fence

import (
	"fmt"
	"math"
	"reflect"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/proven"
)

// CheckFields refuses a column that can draw a value its field in struct type t cannot hold.
func CheckFields(label string, t reflect.Type, columns []tree.RecordColumn) error {
	proof := &valueProof{}
	for _, c := range columns {
		sf, _ := t.FieldByName(c.Name())
		if err := proof.checkField(label+"."+c.Name(), sf.Type, c.Field()); err != nil {
			return err
		}
	}
	return nil
}

// checkField rejects a column a field of Go type ft cannot fill: a datatype other than the one the
// Go type sets, a null outside a pointer, or a value its kind's datatype or range refuses.
func (p *valueProof) checkField(label string, ft reflect.Type, column tree.Node) error {
	elem := ft
	if ft.Kind() == reflect.Pointer {
		elem = ft.Elem()
	}
	kind := FieldKinds[elem.Kind()]
	items, _ := load.ColumnItems(column)
	for _, it := range items {
		if it.DataType() != nil && *it.DataType() != kind.Datatype {
			return fmt.Errorf("%s: its Go type %s sets the datatype %s; drop \"datatype\"", label, ft, kind.Datatype)
		}
	}
	if ft.Kind() != reflect.Pointer && p.proveColumn(column).Nullable {
		return fmt.Errorf("%s: its tag can draw null, which %s cannot hold; make it *%s", label, ft, ft)
	}
	if kind.Datatype == datatype.String {
		return nil
	}
	for _, it := range items {
		v := p.proveColumnItem(it)
		if reason := v.Not[kind.Datatype]; reason != "" {
			return fmt.Errorf("%s (%s): %s", label, ft, reason)
		}
		if !kind.holds(v) {
			return fmt.Errorf("%s (%s): %q is not proven within %s; narrow it to that range, or make the field %s", label, ft, it.Format(), elem.Kind(), kind.wider)
		}
	}
	return nil
}

// FieldKind is what a field of one Go kind holds: the datatype its text proves as, the range a
// number of it stays in, and the kind to name when a value is not proven within that range.
type FieldKind struct {
	Datatype datatype.DataType
	lo, hi   float64
	wider    reflect.Kind
}

// FieldKinds holds each Go kind a fake tag fills, and what a field of it holds.
var FieldKinds = map[reflect.Kind]FieldKind{
	reflect.Bool:    {Datatype: datatype.Boolean},
	reflect.Float32: {datatype.Number, -math.MaxFloat32, math.MaxFloat32, reflect.Float64},
	reflect.Float64: {datatype.Number, -math.MaxFloat64, math.MaxFloat64, reflect.Float64},
	reflect.Int:     {datatype.Integer, math.MinInt, math.MaxInt, reflect.Int64},
	reflect.Int16:   {datatype.Integer, math.MinInt16, math.MaxInt16, reflect.Int64},
	reflect.Int32:   {datatype.Integer, math.MinInt32, math.MaxInt32, reflect.Int64},
	reflect.Int64:   {datatype.Integer, math.MinInt64, math.MaxInt64, reflect.Int64},
	reflect.Int8:    {datatype.Integer, math.MinInt8, math.MaxInt8, reflect.Int64},
	reflect.String:  {Datatype: datatype.String},
	reflect.Uint:    {datatype.Integer, 0, math.MaxUint, reflect.Int64},
	reflect.Uint16:  {datatype.Integer, 0, math.MaxUint16, reflect.Int64},
	reflect.Uint32:  {datatype.Integer, 0, math.MaxUint32, reflect.Int64},
	reflect.Uint64:  {datatype.Integer, 0, math.MaxUint64, reflect.Int64},
	reflect.Uint8:   {datatype.Integer, 0, math.MaxUint8, reflect.Int64},
}

// holds reports whether a field of this kind holds every value v proves. An integer prints
// whole, so its bounds round inward first.
func (k FieldKind) holds(v proven.Facts) bool {
	switch k.Datatype {
	case datatype.String, datatype.Boolean:
		return true
	case datatype.Integer:
		return math.Ceil(v.Lo) >= k.lo && math.Floor(v.Hi) <= k.hi
	}
	return v.Lo >= k.lo && v.Hi <= k.hi
}
