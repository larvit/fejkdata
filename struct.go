package fejkdata

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// FakeStruct fills the struct v points to. Each exported field tagged `fake:"…"` is a
// column of one record: its tag a path or an inline template, told apart as [IsTemplate]
// tells them, and its Go type the column's datatype. A struct field, or a pointer to one,
// fills from its own tags as a record of its own. The first call for a type compiles its
// tags, so a later call for that type fails only as the first did.
func (f *Generator) FakeStruct(v any) error {
	p := reflect.ValueOf(v)
	if p.Kind() != reflect.Pointer || p.IsNil() || p.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("fejkdata: FakeStruct fills a struct through a non-nil pointer, got %T", v)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	shape, err := f.structShapeOf(p.Elem().Type())
	if err != nil {
		return fmt.Errorf("fejkdata: %w", err)
	}
	shape.fill(f.rand, p.Elem())
	return nil
}

// structResult is what compiling a struct type settled: its shape, or why it cannot be filled.
type structResult struct {
	shape *structShape
	err   error
}

// structShapeOf compiles a struct type once and remembers the answer. Callers hold the
// generator's lock.
func (f *Generator) structShapeOf(t reflect.Type) (*structShape, error) {
	if r, done := f.structs[t]; done {
		return r.shape, r.err
	}
	shape, err := compileStruct(f.categories, t, t.String(), map[reflect.Type]bool{})
	if err == nil && shape.empty() {
		err = fmt.Errorf("%s has no fake tags, so nothing to fill", t)
	}
	if f.structs == nil {
		f.structs = map[reflect.Type]structResult{}
	}
	f.structs[t] = structResult{shape, err}
	return shape, err
}

// structShape is a struct type compiled to fill: its tagged fields as one record, the field
// index each column fills, and the struct fields carrying tags of their own.
type structShape struct {
	record  *template
	columns []Column
	fields  []int
	nested  []nestedStruct
}

// nestedStruct is a struct field, or a pointer to one, filled as a record of its own.
type nestedStruct struct {
	index int
	shape *structShape
}

func (s *structShape) empty() bool { return s.record == nil && len(s.nested) == 0 }

// compileStruct compiles struct type t, naming its fields from label. visiting holds the types
// compiling above t, so a pointer back to one is left alone rather than filled without end.
func compileStruct(root map[string]node, t reflect.Type, label string, visiting map[reflect.Type]bool) (*structShape, error) {
	visiting[t] = true
	defer delete(visiting, t)
	shape := &structShape{}
	tags := map[string]any{}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag, tagged := sf.Tag.Lookup("fake")
		if !tagged {
			if err := shape.addNested(root, sf, label, visiting); err != nil {
				return nil, err
			}
			continue
		}
		v, err := tagValue(sf, tag)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", label, sf.Name, err)
		}
		tags[sf.Name] = v
	}
	if len(tags) > 0 {
		if err := shape.compileRecord(root, t, label, tags); err != nil {
			return nil, err
		}
	}
	return shape, nil
}

// addNested adds an untagged exported struct field, or a pointer to one, that carries tags.
func (s *structShape) addNested(root map[string]node, sf reflect.StructField, label string, visiting map[reflect.Type]bool) error {
	t := sf.Type
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if !sf.IsExported() || t.Kind() != reflect.Struct || visiting[t] {
		return nil
	}
	nested, err := compileStruct(root, t, label+"."+sf.Name, visiting)
	if err != nil || nested.empty() {
		return err
	}
	s.nested = append(s.nested, nestedStruct{sf.Index[0], nested})
	return nil
}

// tagValue reads a field's fake tag as the value its column compiles from: an inline template
// as written, or a path as the reference {/path}.
func tagValue(sf reflect.StructField, tag string) (any, error) {
	if err := checkTaggedType(sf); err != nil {
		return nil, err
	}
	inline, err := isTemplate(tag)
	if err != nil {
		return nil, err
	}
	if !inline {
		for _, seg := range strings.Split(tag, ".") {
			if err := checkName(seg); err != nil {
				return nil, fmt.Errorf("path %w", err)
			}
		}
		return "{/" + tag + "}", nil
	}
	v, err := inputValue(tag)
	if s, isString := v.(string); isString && isLoneReference(s) {
		path := s[2 : len(s)-1]
		return nil, fmt.Errorf("%s is the path %s written as a template; write fake:%q", s, path, path)
	}
	return v, err
}

// isLoneReference reports whether a format is one {/path} token alone, which a path tag spells.
func isLoneReference(format string) bool {
	return len(format) > len("{/}") && strings.HasPrefix(format, "{/") && strings.HasSuffix(format, "}") &&
		!strings.ContainsAny(format[1:len(format)-1], "{}|(")
}

// checkTaggedType rejects a tagged field no column can fill.
func checkTaggedType(sf reflect.StructField) error {
	elem := sf.Type
	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}
	_, holds := columnKinds[elem.Kind()]
	switch {
	case !sf.IsExported():
		return errors.New("unexported, so its fake tag cannot fill it")
	case holds:
		return nil
	case elem.Kind() == reflect.Struct:
		return errors.New("a struct field fills from the tags on its own fields; drop this one")
	}
	return fmt.Errorf("a fake tag fills a string, bool, integer or float field, or a pointer to one, not %s", sf.Type)
}

// compileRecord compiles the tagged fields of t as one record, and proves each column holds
// only what its field's Go type can.
func (s *structShape) compileRecord(root map[string]node, t reflect.Type, label string, tags map[string]any) error {
	tags["format"] = ""
	n, err := compile(tags)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err := bindInline(n, label, root); err != nil {
		return err
	}
	record, columns, err := recordOf(n)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	proof := &valueProof{}
	s.fields = make([]int, len(columns))
	for i, c := range columns {
		sf, _ := t.FieldByName(c.Name)
		if c.DataType != DataTypeString {
			return fmt.Errorf("%s.%s: its Go type %s sets the datatype; drop \"datatype\"", label, c.Name, sf.Type)
		}
		if err := proof.checkField(label+"."+c.Name, sf.Type, record.fields[c.Name]); err != nil {
			return err
		}
		s.fields[i] = sf.Index[0]
	}
	s.record, s.columns = record, columns
	return nil
}

// columnKind is what a field of one Go kind holds: the datatype its text proves as, and the
// range its value stays in.
type columnKind struct {
	datatype DataType
	lo, hi   float64
}

var columnKinds = map[reflect.Kind]columnKind{
	reflect.Bool:    {DataTypeBoolean, -math.MaxFloat64, math.MaxFloat64},
	reflect.Float32: {DataTypeNumber, -math.MaxFloat32, math.MaxFloat32},
	reflect.Float64: {DataTypeNumber, -math.MaxFloat64, math.MaxFloat64},
	reflect.Int:     {DataTypeInteger, math.MinInt, math.MaxInt},
	reflect.Int16:   {DataTypeInteger, math.MinInt16, math.MaxInt16},
	reflect.Int32:   {DataTypeInteger, math.MinInt32, math.MaxInt32},
	reflect.Int64:   {DataTypeInteger, math.MinInt64, math.MaxInt64},
	reflect.Int8:    {DataTypeInteger, math.MinInt8, math.MaxInt8},
	reflect.String:  {DataTypeString, -math.MaxFloat64, math.MaxFloat64},
	reflect.Uint:    {DataTypeInteger, 0, math.MaxUint},
	reflect.Uint16:  {DataTypeInteger, 0, math.MaxUint16},
	reflect.Uint32:  {DataTypeInteger, 0, math.MaxUint32},
	reflect.Uint64:  {DataTypeInteger, 0, math.MaxUint64},
	reflect.Uint8:   {DataTypeInteger, 0, math.MaxUint8},
}

// checkField rejects a column some render of which a field of Go type ft cannot hold: a null
// outside a pointer, or a value its kind's datatype or range refuses.
func (p *valueProof) checkField(label string, ft reflect.Type, column node) error {
	items, nullable := columnItems(column)
	elem := ft
	if ft.Kind() == reflect.Pointer {
		elem = ft.Elem()
	} else if nullable {
		return fmt.Errorf("%s: its tag can draw null, which %s cannot hold; make it *%s", label, ft, ft)
	}
	kind := columnKinds[elem.Kind()]
	if kind.datatype == DataTypeString {
		return nil
	}
	for _, it := range items {
		v := p.of(it)
		reason := v.not[kind.datatype]
		if reason == "" && (v.lo < kind.lo || v.hi > kind.hi) {
			reason = fmt.Sprintf("%q is not proven within %s", it.format, elem.Kind())
		}
		if reason != "" {
			return fmt.Errorf("%s (%s): %s", label, ft, reason)
		}
	}
	return nil
}

// fill draws the record into v's tagged fields, then each nested struct as a record of its own.
func (s *structShape) fill(sess *session, v reflect.Value) {
	if s.record != nil {
		for i, c := range renderRecord(sess, s.record, s.columns).columns {
			setColumn(v.Field(s.fields[i]), c)
		}
	}
	for _, n := range s.nested {
		field := v.Field(n.index)
		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		n.shape.fill(sess, field)
	}
}

// setColumn writes a drawn column into its field: a null as a nil pointer, a value through a
// fresh pointer or straight into the field.
func setColumn(field reflect.Value, c Column) {
	if field.Kind() != reflect.Pointer {
		setText(field, c.Value)
		return
	}
	if c.Null {
		field.SetZero()
		return
	}
	value := reflect.New(field.Type().Elem())
	setText(value.Elem(), c.Value)
	field.Set(value)
}

// setText parses text into a field of one of columnKinds, which checkField proved it parses as.
func setText(field reflect.Value, text string) {
	var err error
	switch kind := columnKinds[field.Kind()]; {
	case kind.datatype == DataTypeString:
		field.SetString(text)
	case kind.datatype == DataTypeBoolean:
		var b bool
		b, err = strconv.ParseBool(text)
		field.SetBool(b)
	case kind.datatype == DataTypeNumber:
		var x float64
		x, err = strconv.ParseFloat(text, field.Type().Bits())
		field.SetFloat(x)
	case field.CanInt():
		var n int64
		n, err = strconv.ParseInt(text, 10, field.Type().Bits())
		field.SetInt(n)
	default:
		var n uint64
		n, err = strconv.ParseUint(text, 10, field.Type().Bits())
		field.SetUint(n)
	}
	if err != nil {
		panic(fmt.Sprintf("fejkdata: %q reached a %s field unproven: %v", text, field.Type(), err))
	}
}
