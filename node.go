package fejkdata

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

// node is a compiled element of the namespace tree: a folder, choice, null,
// template, table, column or row. Compiling JSON into these once (see compile)
// means rendering never re-inspects the raw JSON or re-sums weights.
type node interface{ isNode() }

// folder is a namespace of named children, built from a directory of JSON files
// and subdirectories. It has no value of its own: descend into a named child by
// dot path; rendering one is an error (see Fake).
type folder struct{ children map[string]node }

func (*folder) isNode() {}

// choice picks one of its items. cum holds cumulative weights for a weighted
// pick; when nil the choice is uniform and selection is O(1). shared is the set of
// relative dot paths every item can address, so descend and List both read the one
// answer to what a path may reach through this choice.
type choice struct {
	items  []node
	cum    []float64
	shared map[string]bool
}

func (*choice) isNode() {}

// null is a column's missing value, rendered ""; sized so two nulls are two map keys.
type null struct{ _ byte }

func (*null) isNode() {}

// template renders a format string, substituting {tokens} from fields. A bare
// JSON string is a template with no fields. repeat (default 1) renders that format
// that many times and joins the results with separator (default ""), each render
// an independent pick. A format holding a reference compiles in `linkTemplateRefs`.
type template struct {
	// Filled by `compileString`, `compileTemplate` and `table.compileWhole`:
	format     string
	tokens     []formatToken
	fields     map[string]node
	repeat     int
	separator  string
	datatype   DataType
	drawGroup  string // the draw group it draws in, as written; "" keeps its caller's
	fromString bool   // written as a JSON string rather than an object
	record     bool   // compiled at the top without a repeat, so its fields are record columns
	table      *table // the table whose format this is, whose columns are the fields

	// Filled by `checkCells`, after the cell's own format has compiled:
	cellOf  *table // the table whose cell this is
	cellRow int    // the row the cell sits in

	// Filled by `template.compileFormat`, from `template.format` and `template.refs`:
	ops   []op   // what expand walks
	grow  int    // minimum output size, to size the render buffer
	fixed bool   // no op varies, so every render is lit
	lit   string // the whole output when fixed
	// bound maps each field the format addresses by dotted path to one path token
	// reading it, which is the half of an overlap the fences name. nil when the
	// format takes no path.
	bound map[string]string
	// held is every name drawn once per expansion: the bound levels above, plus the
	// siblings a {calc()} reads. nil when the format holds nothing (see expand).
	held      map[string]bool
	heldLocal bool // some held name is kept by the expansion itself, so expand makes its hold

	// Filled by `linkTemplateRefs` and `keyDrawGroup`, from the assembled tree:
	refs         map[string]refBinding // each reference the format reads -> what it is bound to
	refHeads     map[string]node       // each refBinding.key -> the category it names
	readsColumn  *columnRead           // set when the format is one reference alone reading a record's column
	drawGroupKey string                // its draw group keyed by its category: what a render reads its reference paths under
}

func (*template) isNode() {}

// head is the node an arm's key names: a sibling field, or a reference's category.
func (t *template) head(key string) node {
	if isRef(key) {
		return t.refHeads[key]
	}
	return t.fields[key]
}

// compile converts parsed JSON — a category or an inline template — into a node tree,
// validating structure up front.
func compile(v any) (node, error) {
	return compileAt(v, atTop)
}

// compileCategory compiles a data file's value, which may be a table over a rows
// file beside it, and refuses a choice that is a table written as templates.
func compileCategory(v any, name string, files *categoryFiles) (node, error) {
	if m, ok := v.(map[string]any); ok {
		if _, isTable := m["rows"]; isTable {
			return compileTable(m, name, files)
		}
	}
	if items, ok := v.([]any); ok {
		if err := checkNotRows(items, name); err != nil {
			return nil, err
		}
	}
	return compile(v)
}

// checkNotRows refuses a choice of templates sharing one format and one set of
// string fields: each item is a row, and the rows file is the spelling for that.
// docs/decisions.md#the-choice-of-rows-fence-guards-a-data-files-root-and-requires-string-fields
func checkNotRows(items []any, name string) error {
	var format string
	var fields []string
	weighted := false
	for i, raw := range items {
		f, keys, w, isRow := rowShape(raw)
		if !isRow || i > 0 && (f != format || !slices.Equal(keys, fields)) {
			return nil
		}
		format, fields, weighted = f, keys, weighted || w
	}
	if len(items) < 2 || len(fields) == 0 {
		return nil
	}
	weight := ""
	if weighted {
		weight = `, a weight column, and "weight" naming it`
	}
	return fmt.Errorf("a choice of %d templates with one format and the fields %v is a table; write the rows in %s.tsv with the header %q%s, and the category as {\"format\": %q, \"rows\": \"%s.tsv\"}",
		len(items), fields, name, strings.Join(fields, "\t"), weight, format, name)
}

// rowShape reads a choice item as a row: a template whose every field is a string,
// with its format, its sorted field names, and whether it carries a weight.
func rowShape(raw any) (format string, fields []string, weighted, isRow bool) {
	m, ok := raw.(map[string]any)
	if !ok {
		return "", nil, false, false
	}
	format, _ = m["format"].(string)
	for k, v := range m {
		if k == "weight" {
			weighted = true
			continue
		}
		if _, isString := v.(string); !isString || isOption(k) && k != "format" {
			return "", nil, false, false
		}
		if k != "format" {
			fields = append(fields, k)
		}
	}
	sort.Strings(fields)
	return format, fields, weighted, true
}

// compileAt compiles a node that is no choice's item. Only a choice's items carry a
// weight, so one here would be inert whatever its type.
func compileAt(v any, pos position) (node, error) {
	if m, ok := v.(map[string]any); ok {
		if _, weighted := m["weight"]; weighted {
			return nil, fmt.Errorf("weight only skews a choice's items, so it has no effect here; it is an option and can never be a field")
		}
	}
	return compileItem(v, pos)
}

// compileItem compiles one node, allowing the weight a choice item may carry.
func compileItem(v any, pos position) (node, error) {
	switch v := v.(type) {
	case string:
		return compileString(v)
	case []any:
		return compileChoice(v, pos)
	case map[string]any:
		return compileTemplate(v, pos)
	case nil:
		if pos != inColumn {
			return nil, fmt.Errorf(`null is a record column's value; here it only renders "", so write ""`)
		}
		return &null{}, nil
	default:
		return nil, fmt.Errorf("a template value must be a string, a list or an object, not %s", jsonKind(v))
	}
}

// jsonKind names a JSON value a template cannot hold, in the data format's own
// terms rather than the decoding library's.
func jsonKind(v any) string {
	switch v.(type) {
	case float64:
		return "a number"
	case bool:
		return "a boolean"
	}
	return fmt.Sprintf("%T", v)
}

func compileString(s string) (node, error) {
	toks, err := parseFormat(s)
	if err != nil {
		return nil, err
	}
	if err := checkTokens(toks, nil); err != nil {
		return nil, err
	}
	t := &template{format: s, tokens: toks, repeat: 1, fromString: true}
	if err := t.compileRefFree(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *template) compileRefFree() error {
	if len(refTokens(t.tokens)) > 0 {
		return nil
	}
	return t.compileFormat()
}

// compileFormat compiles the format into ops, and applies the fences that need the
// compiled reads.
func (t *template) compileFormat() error {
	c := compileOps(t.tokens, t.refs)
	t.ops, t.grow, t.bound, t.held, t.heldLocal = c.ops, c.grow, c.bound, c.held, c.heldLocal
	t.fixed = true
	for _, o := range t.ops {
		if o.kind != 'l' {
			t.fixed = false
		}
	}
	if t.fixed && len(t.ops) == 1 {
		t.lit = t.ops[0].lit
	}
	if err := checkNoOverlap(t.ops, t.bound); err != nil {
		return err
	}
	return checkNoRepeatedRead(c)
}

func compileChoice(items []any, pos position) (node, error) {
	itemPos := inFormat
	if pos == inColumn {
		itemPos = inColumn
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("empty choice")
	}
	if len(items) == 1 {
		return nil, fmt.Errorf("a one-item choice is its item; write the item")
	}
	if err := checkNoRepeatedItem(items); err != nil {
		return nil, err
	}
	c := &choice{items: make([]node, len(items))}
	cum := make([]float64, len(items))
	var total float64
	weighted := false
	for i, raw := range items {
		w, err := weightOf(raw)
		if err != nil {
			return nil, err
		}
		if w != 1 {
			weighted = true
		}
		total += w
		cum[i] = total
		n, err := compileItem(raw, itemPos)
		if err != nil {
			return nil, err
		}
		c.items[i] = n
	}
	if weighted { // uniform choices skip the weight table and pick in O(1)
		if math.IsInf(total, 1) {
			return nil, fmt.Errorf("choice weights must sum to a finite number, got %v", total)
		}
		c.cum = cum
	}
	c.shared = sharedPaths(c.items)
	return c, nil
}

// checkNoRepeatedItem rejects a choice that lists one item twice: a pick is even
// over the items, so a repeat is a second spelling of weight. The error names the
// spelling that does skew a pick.
func checkNoRepeatedItem(items []any) error {
	seen := make(map[string]int, len(items))
	for i, raw := range items {
		key, isString := raw.(string)
		if !isString {
			b, err := json.Marshal(raw)
			if err != nil {
				return err
			}
			key = "\x00" + string(b)
		}
		if j, dup := seen[key]; dup {
			if s, isString := raw.(string); isString {
				return fmt.Errorf("choice item %q is repeated; skew the odds with a weight instead: { \"format\": %q, \"weight\": 2 }", s, s)
			}
			if raw == nil {
				return fmt.Errorf("choice item %d repeats null; a null takes no weight, so weight the other items instead", i)
			}
			return fmt.Errorf("choice item %d repeats item %d; skew the odds with a weight on one of them instead", i, j)
		}
		seen[string(key)] = i
	}
	return nil
}

func compileTemplate(m map[string]any, pos position) (node, error) {
	if _, isTable := m["rows"]; isTable {
		return nil, fmt.Errorf("rows names a TSV beside a category's file, so only a category is a table; an inline template has no file beside it")
	}
	o, err := readOptions(m, pos)
	if err != nil {
		return nil, err
	}
	fieldPos := inFormat
	if pos == atTop && o.repeat == 1 {
		fieldPos = inColumn
	}
	fields, err := compileFields(m, fieldPos)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 && o.repeat == 1 && !o.weighted && o.datatype == DataTypeString && o.group == "" {
		return nil, fmt.Errorf("an object holding only a format is a string; write %q", o.format)
	}
	toks, err := parseFormat(o.format)
	if err != nil {
		return nil, err
	}
	if err := checkTokens(toks, fields); err != nil {
		return nil, err
	}
	if err := checkNestedDrawGroup(fields, o.group); err != nil {
		return nil, err
	}
	t := &template{format: o.format, tokens: toks, fields: fields, repeat: o.repeat, separator: o.separator, datatype: o.datatype, drawGroup: o.group, record: fieldPos == inColumn}
	if err := t.compileRefFree(); err != nil {
		return nil, err
	}
	return t, nil
}

// templateOptions is what a template object's option keys say.
type templateOptions struct {
	datatype  DataType
	format    string
	group     string
	repeat    int
	separator string
	weighted  bool
}

func readOptions(m map[string]any, pos position) (templateOptions, error) {
	var o templateOptions
	format, ok := m["format"].(string)
	if !ok {
		return o, fmt.Errorf("template object missing string \"format\"")
	}
	o.format = format
	repeat, err := repeatOf(m)
	if err != nil {
		return o, err
	}
	o.repeat = repeat
	if o.datatype, err = datatypeOf(m, pos); err != nil {
		return o, err
	}
	if o.group, err = drawGroupOf(m, repeat); err != nil {
		return o, err
	}
	if sv, ok := m["separator"]; ok {
		if o.separator, ok = sv.(string); !ok {
			return o, fmt.Errorf("separator must be a string, got %T", sv)
		}
		if repeat == 1 {
			return o, fmt.Errorf("separator joins repeated renders, so it has no effect without a repeat above 1")
		}
		if o.separator == "" {
			return o, fmt.Errorf("separator \"\" is the default, so it has no effect; drop it")
		}
	}
	_, o.weighted = m["weight"]
	return o, nil
}

// compileFields compiles every non-option key of a template object, in name order
// so which of several bad fields is reported does not vary.
func compileFields(m map[string]any, pos position) (map[string]node, error) {
	fields := make(map[string]node, len(m))
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if isOption(k) {
			continue
		}
		if err := checkName(k); err != nil {
			return nil, fmt.Errorf("field %w", err)
		}
		n, err := compileAt(m[k], pos)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", k, err)
		}
		fields[k] = n
	}
	return fields, nil
}

// repeatOf reads a template's "repeat" (default 1): how many times its format
// is rendered and concatenated. A present one must be an integer above 1.
func repeatOf(m map[string]any) (int, error) {
	rv, ok := m["repeat"]
	if !ok {
		return 1, nil
	}
	r, ok := rv.(float64)
	if !ok {
		return 0, fmt.Errorf("repeat must be a number, got %T", rv)
	}
	if math.IsNaN(r) || math.IsInf(r, 0) || r < 1 || r != math.Trunc(r) {
		return 0, fmt.Errorf("repeat must be a positive integer, got %v", rv)
	}
	if r == 1 {
		return 0, fmt.Errorf("repeat 1 is the default, so it has no effect; drop it")
	}
	if r > MaxRepeat { // caps the renders one repeat asks for; repeatCheck bounds what nested ones multiply to
		return 0, fmt.Errorf("repeat %v exceeds the maximum %d", rv, MaxRepeat)
	}
	return int(r), nil
}

// drawGroupOf reads a template's "drawGroup" (default ""), which a repeat cannot carry: each
// iteration renders in no draw group.
func drawGroupOf(m map[string]any, repeat int) (string, error) {
	v, ok := m["drawGroup"]
	if !ok {
		return "", nil
	}
	name, ok := v.(string)
	switch {
	case !ok:
		return "", fmt.Errorf("drawGroup must be a string, got %T", v)
	case name == "":
		return "", fmt.Errorf(`drawGroup "" is the default, so it has no effect; drop it`)
	case repeat > 1:
		return "", fmt.Errorf("drawGroup %q on a repeat names nothing, since each iteration is a render of its own; drop it", name)
	}
	return name, nil
}

// keyDrawGroup keys t's draw group by the category t sits in, "" for an inline template, so a name
// is local to its category.
func (t *template) keyDrawGroup(category string) {
	if t.drawGroup != "" {
		t.drawGroupKey = category + "/" + t.drawGroup
	}
}

// weightOf reads a node's "weight" (default 1) from its raw JSON form. Only
// template objects carry weight; a present one must be finite and positive.
func weightOf(raw any) (float64, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return 1, nil
	}
	wv, ok := m["weight"]
	if !ok {
		return 1, nil
	}
	w, ok := wv.(float64)
	if !ok {
		return 0, fmt.Errorf("weight must be a number, got %T", wv)
	}
	if w < 0 || math.IsNaN(w) || math.IsInf(w, 0) {
		return 0, fmt.Errorf("weight must be finite and positive, got %v", w)
	}
	if w == 0 {
		return 0, fmt.Errorf("weight 0 means the item is never drawn; remove the item instead")
	}
	if w == 1 {
		return 0, fmt.Errorf("weight 1 is the default, so it has no effect; drop it")
	}
	return w, nil
}

// reservedInName is what a category, folder or field name may not contain: a dot
// separates the segments of a path, '|' the arms of a token, '(' opens a function
// call, braces delimit the token, '/' starts a reference, and brackets and a quote
// open a JSON value. A name carrying one is rejected where it is authored rather
// than where it would be unreachable.
const reservedInName = ".|({}/[]\""

// reservedList spells reservedInName for an error message, so the two cannot drift.
var reservedList = strings.Join(strings.Split(reservedInName, ""), " ")

// checkName rejects a name the dot path, {token} and JSON grammars cannot spell, or a struct
// tag cannot read.
// Both a category or folder and a field go through it, so there is one answer to
// what a name may contain.
func checkName(name string) error {
	if name == "" {
		return fmt.Errorf("%q is empty, which is not a path segment, so List never offers it", name)
	}
	if name == "-" {
		return fmt.Errorf(`%q is reserved: the struct tag fake:"-" leaves a field unfilled, so no tag could read it; rename it`, name)
	}
	if i := strings.IndexAny(name, reservedInName); i >= 0 {
		return fmt.Errorf("%q contains %q; a name may not use %s, which the dot path, {token} and JSON grammars reserve",
			name, name[i:i+1], reservedList)
	}
	return nil
}

// checkPathNames rejects a dotted path with a segment no name may be.
func checkPathNames(path string) error {
	segs, err := splitPath(path)
	if err != nil {
		return err
	}
	for _, seg := range names(segs) {
		if err := checkName(seg); err != nil {
			return fmt.Errorf("path %w", err)
		}
	}
	return nil
}

// isOption reports whether a template key configures the node instead of naming a
// field. These names can never be fields.
func isOption(name string) bool {
	switch name {
	case "datatype", "drawGroup", "format", "repeat", "separator", "weight":
		return true
	}
	return false
}
