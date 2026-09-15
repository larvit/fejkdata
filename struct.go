package fejkdata

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// FakeStruct fills the struct v points to. Each exported field tagged `fake:"…"` is a
// column of one record: its tag a path or an inline template, told apart as [IsTemplate]
// tells them, and its Go type the column's datatype. The fields an embedded struct promotes
// are columns of that record too, while a named struct field, or a pointer to one, fills
// from its own tags as a record of its own. The first call for a type compiles its tags,
// so a later call for that type fails only as the first did.
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

type structResult struct {
	shape *structShape
	err   error
}

// maxStructs caps the structs compiling one type walks through its fields.
const maxStructs = 1 << 10

// structShapeOf compiles a struct type once and remembers the answer. Callers hold the
// generator's lock.
func (f *Generator) structShapeOf(t reflect.Type) (*structShape, error) {
	if r, done := f.structs[t]; done {
		return r.shape, r.err
	}
	label := t.Name()
	if label == "" {
		label = "struct"
	}
	sc := &structCompile{root: f.categories, visiting: map[reflect.Type]bool{}, structs: maxStructs}
	shape, err := sc.record(t, label)
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
// index path each column fills, and the struct fields carrying tags of their own.
type structShape struct {
	record  *template
	columns []Column
	fields  [][]int
	nested  []nestedStruct
}

// nestedStruct is a named struct field, or a pointer to one, filled as a record of its own.
type nestedStruct struct {
	index []int
	shape *structShape
}

func (s *structShape) empty() bool { return s.record == nil && len(s.nested) == 0 }

// structCompile is what compiling one struct type shares across the structs it reaches: the
// loaded tree, the types compiling or embedded above, so a pointer back to one is left alone
// rather than filled without end, and how many more structs it may walk.
type structCompile struct {
	root     map[string]node
	visiting map[reflect.Type]bool
	structs  int
}

// structFields gathers what one struct type fills: its tagged fields, those its embedded
// structs promote included, as the tags of one record, and its named struct fields as nested
// records.
type structFields struct {
	*structCompile
	t     reflect.Type
	label string
	tags  map[string]any
	shape *structShape
}

func (sc *structCompile) record(t reflect.Type, label string) (*structShape, error) {
	if err := sc.spend(label); err != nil {
		return nil, err
	}
	sc.visiting[t] = true
	defer delete(sc.visiting, t)
	c := &structFields{structCompile: sc, t: t, label: label, tags: map[string]any{}, shape: &structShape{}}
	if err := c.walk(t, nil); err != nil {
		return nil, err
	}
	if len(c.tags) > 0 {
		if err := c.shape.compileRecord(sc.root, t, label, c.tags); err != nil {
			return nil, err
		}
	}
	return c.shape, nil
}

func (sc *structCompile) spend(label string) error {
	if sc.structs--; sc.structs >= 0 {
		return nil
	}
	return fmt.Errorf(`%s: the struct fields reach more than %d structs; leave a struct field unfilled with fake:"-"`, label, maxStructs)
}

// walk gathers the fields of struct type t, which sits at index within c.t.
func (c *structFields) walk(t reflect.Type, index []int) error {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		sf.Index = append(index[:len(index):len(index)], i)
		if err := c.field(sf); err != nil {
			return err
		}
	}
	return nil
}

func (c *structFields) field(sf reflect.StructField) error {
	tag, tagged := sf.Tag.Lookup("fake")
	elem := structOf(sf.Type)
	switch {
	case tagged && tag == "-":
		if elem == nil {
			return fmt.Errorf(`%s.%s: fake:"-" leaves a struct field unfilled, and any other untagged field keeps its value already; drop the tag`, c.label, sf.Name)
		}
		return nil
	case tagged:
		return c.column(sf, tag)
	case elem == nil || c.visiting[elem]:
		return nil
	case sf.Anonymous:
		return c.embed(sf, elem)
	case sf.IsExported():
		return c.nest(sf, elem)
	}
	return nil
}

// structOf is the struct type a field holds, by value or through a pointer; nil when none.
func structOf(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	return t
}

// column adds a tagged field's tag to the record, refusing one another field hides.
func (c *structFields) column(sf reflect.StructField, tag string) error {
	if visible, ok := c.t.FieldByName(sf.Name); !ok || !slices.Equal(visible.Index, sf.Index) {
		return fmt.Errorf("%s.%s: hidden by another field named %s, so its fake tag cannot fill it; rename one", c.label, fieldPath(c.t, sf.Index), sf.Name)
	}
	v, err := tagValue(sf, tag)
	if err != nil {
		return fmt.Errorf("%s.%s: %w", c.label, sf.Name, err)
	}
	c.tags[sf.Name] = v
	return nil
}

// fieldPath names the field at index within t through each struct it is embedded in.
func fieldPath(t reflect.Type, index []int) string {
	names := make([]string, len(index))
	for i := range index {
		names[i] = t.FieldByIndex(index[:i+1]).Name
	}
	return strings.Join(names, ".")
}

func (c *structFields) embed(sf reflect.StructField, elem reflect.Type) error {
	if err := c.spend(c.label + "." + fieldPath(c.t, sf.Index)); err != nil {
		return err
	}
	c.visiting[elem] = true
	defer delete(c.visiting, elem)
	tags, nested := len(c.tags), len(c.shape.nested)
	if err := c.walk(elem, sf.Index); err != nil {
		return err
	}
	if sf.Type.Kind() == reflect.Pointer && !sf.IsExported() && (len(c.tags) > tags || len(c.shape.nested) > nested) {
		return fmt.Errorf("%s.%s: an unexported embedded pointer field cannot be set, so the tags beneath it cannot fill; embed %s by value", c.label, fieldPath(c.t, sf.Index), elem)
	}
	return nil
}

func (c *structFields) nest(sf reflect.StructField, elem reflect.Type) error {
	nested, err := c.record(elem, c.label+"."+sf.Name)
	if err != nil || nested.empty() {
		return err
	}
	c.shape.nested = append(c.shape.nested, nestedStruct{sf.Index, nested})
	return nil
}

// tagValue reads a field's fake tag as the value its column compiles from: an inline template
// as written, or a path as the reference {/path}.
func tagValue(sf reflect.StructField, tag string) (any, error) {
	if err := checkTaggedType(sf); err != nil {
		return nil, err
	}
	inline, err := isTemplate(tag)
	switch {
	case err != nil:
		return nil, err
	case inline:
		return inputValue(tag)
	}
	if err := checkPathNames(tag); err != nil {
		return nil, err
	}
	return "{/" + tag + "}", nil
}

// checkTaggedType rejects a tagged field no column can fill.
func checkTaggedType(sf reflect.StructField) error {
	_, holds := columnKinds[sf.Type.Kind()]
	if sf.Type.Kind() == reflect.Pointer {
		_, holds = columnKinds[sf.Type.Elem().Kind()]
	}
	switch {
	case !sf.IsExported():
		return errors.New("unexported, so its fake tag cannot fill it")
	case holds:
		return nil
	case structOf(sf.Type) != nil:
		return errors.New(`a struct field fills from the tags on its own fields; drop this one, or write fake:"-" to leave it unfilled`)
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
	if err := bindInline(n, label, root, checkRenders); err != nil {
		return err
	}
	record, columns, err := recordOf(n)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	proof := &valueProof{}
	s.fields = make([][]int, len(columns))
	for i, c := range columns {
		sf, _ := t.FieldByName(c.Name)
		if err := proof.checkField(label+"."+c.Name, sf.Type, record.fields[c.Name]); err != nil {
			return err
		}
		s.fields[i] = sf.Index
	}
	s.record, s.columns = record, columns
	return nil
}

// columnKind is what a field of one Go kind holds: the datatype its text proves as, the range a
// number of it stays in, and the kind to name when a value is not proven within that range.
type columnKind struct {
	datatype DataType
	lo, hi   float64
	wider    reflect.Kind
}

var columnKinds = map[reflect.Kind]columnKind{
	reflect.Bool:    {datatype: DataTypeBoolean},
	reflect.Float32: {DataTypeNumber, -math.MaxFloat32, math.MaxFloat32, reflect.Float64},
	reflect.Float64: {DataTypeNumber, -math.MaxFloat64, math.MaxFloat64, reflect.Float64},
	reflect.Int:     {DataTypeInteger, math.MinInt, math.MaxInt, reflect.Int64},
	reflect.Int16:   {DataTypeInteger, math.MinInt16, math.MaxInt16, reflect.Int64},
	reflect.Int32:   {DataTypeInteger, math.MinInt32, math.MaxInt32, reflect.Int64},
	reflect.Int64:   {DataTypeInteger, math.MinInt64, math.MaxInt64, reflect.Int64},
	reflect.Int8:    {DataTypeInteger, math.MinInt8, math.MaxInt8, reflect.Int64},
	reflect.String:  {datatype: DataTypeString},
	reflect.Uint:    {DataTypeInteger, 0, math.MaxUint, reflect.Int64},
	reflect.Uint16:  {DataTypeInteger, 0, math.MaxUint16, reflect.Int64},
	reflect.Uint32:  {DataTypeInteger, 0, math.MaxUint32, reflect.Int64},
	reflect.Uint64:  {DataTypeInteger, 0, math.MaxUint64, reflect.Int64},
	reflect.Uint8:   {DataTypeInteger, 0, math.MaxUint8, reflect.Int64},
}

// holds reports whether a field of this kind holds every value v proves. An integer prints
// whole, so its bounds round inward first.
func (k columnKind) holds(v proven) bool {
	switch k.datatype {
	case DataTypeString, DataTypeBoolean:
		return true
	case DataTypeInteger:
		return math.Ceil(v.lo) >= k.lo && math.Floor(v.hi) <= k.hi
	}
	return v.lo >= k.lo && v.hi <= k.hi
}

// checkField rejects a column a field of Go type ft cannot fill: a datatype, which the Go type
// sets, a null outside a pointer, or a value its kind's datatype or range refuses.
func (p *valueProof) checkField(label string, ft reflect.Type, column node) error {
	items, _ := columnItems(column)
	for _, it := range items {
		if it.datatype != DataTypeString {
			return fmt.Errorf("%s: its Go type %s sets the datatype; drop \"datatype\"", label, ft)
		}
	}
	elem := ft
	if ft.Kind() == reflect.Pointer {
		elem = ft.Elem()
	} else if p.column(column).null {
		return fmt.Errorf("%s: its tag can draw null, which %s cannot hold; make it *%s", label, ft, ft)
	}
	kind := columnKinds[elem.Kind()]
	if kind.datatype == DataTypeString {
		return nil
	}
	for _, it := range items {
		v := p.columnItem(it)
		if reason := v.not[kind.datatype]; reason != "" {
			return fmt.Errorf("%s (%s): %s", label, ft, reason)
		}
		if !kind.holds(v) {
			return fmt.Errorf("%s (%s): %q is not proven within %s; make it %s", label, ft, it.format, elem.Kind(), kind.wider)
		}
	}
	return nil
}

// fill draws the record into v's tagged fields, then each nested struct as a record of its own.
func (s *structShape) fill(sess *session, v reflect.Value) {
	if s.record != nil {
		for i, c := range renderRecord(sess, s.record, s.columns).columns {
			setColumn(fieldAt(v, s.fields[i]), c)
		}
	}
	for _, n := range s.nested {
		field := fieldAt(v, n.index)
		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		n.shape.fill(sess, field)
	}
}

// fieldAt is v's field at index, allocating each nil embedded pointer on the way.
func fieldAt(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
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
