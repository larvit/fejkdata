package fejkdata

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/grammar"
)

// node is a compiled element of the namespace tree: a folder, choice, null,
// template, table, column or row.
type node interface{ isNode() }

// folder is a namespace of named children, built from a directory of JSON files
// and subdirectories. It has no value of its own: descend into a named child by
// dot path; rendering one is an error (see Fake).
type folder struct {
	children map[string]node
	unloaded map[string]shippedEntry // the shipped categories here a call has not reached yet
}

func (*folder) isNode() {}

// choice picks one of its items. cum holds cumulative weights for a weighted
// pick; when nil the choice is uniform and selection is O(1). shared is the set of
// relative dot paths every item can address, so carriedByAll and List both read the one
// answer to what a path may reach through this choice.
type choice struct {
	items     []node
	cum       []float64
	shared    map[string]bool
	nameScope *nameScope // where its items look a name up, set by bindNames
}

func (*choice) isNode() {}

// nullItem is a column's missing value, rendered ""; sized so two nulls are two map keys.
type nullItem struct{ _ byte }

func (*nullItem) isNode() {}

// template renders a format string, substituting {tokens} from fields. A bare
// JSON string is a template with no fields. repeat (default 1) renders that format
// that many times and joins the results with separator (default ""), each render
// an independent pick. Every format compiles in `linkTemplate`.
type template struct {
	// Filled by `compileString`, `compileTemplate` and `table.compileRowFormat`:
	format     string
	tokens     []grammar.Token
	fields     map[string]node
	repeat     int
	separator  string
	datatype   DataType
	fromString bool          // written as a JSON string rather than an object
	isRecord   bool          // compiled at the top without a repeat, so its fields are record columns
	unbound    []unboundRead // the heads its tokens read that no field holds

	// Filled by `bindNames`, from the compiled category:
	nameScope    *nameScope // where its tokens look a name up
	ownNameScope *nameScope // the scope it renders a frame of: a category's, on its root, or a repeat's, per iteration

	// Filled by `linkTemplate`, from the assembled tree:
	link     templateLink
	compiled formatOps
}

// templateLink is what a template resolves to in the assembled tree.
type templateLink struct {
	refs        map[string]refBinding // each reference the format reads -> what it resolves to
	refHeads    map[string]node       // each refBinding.head -> the category it names
	readsColumn *columnRead           // set when the format is one reference or name read alone reading a record's column
}

func (*template) isNode() {}

// head is the node an arm's head names: a sibling field, or a reference's category.
func (t *template) head(name string) node {
	if grammar.IsRef(name) {
		return t.link.refHeads[name]
	}
	return t.fields[name]
}

// compile converts parsed JSON — a category or an inline template — into a node tree,
// validating structure up front.
func compile(v any) (node, error) {
	n, err := compileAt(v, atTop)
	if err != nil {
		return nil, err
	}
	return n, bindNames(n)
}

// compileCategory compiles a data file's value, which may be a table over a rows
// file beside it, and refuses a choice that is a table written as templates.
func compileCategory(c datafiles.Category) (node, error) {
	if m, ok := c.JSON.(map[string]any); ok {
		if _, isTable := m["rows"]; isTable {
			return compileTable(m, c.Folders, c.Name, c.ReadRows)
		}
	}
	if items, ok := c.JSON.([]any); ok {
		if err := checkNotRows(items, c.Name); err != nil {
			return nil, err
		}
	}
	return compile(c.JSON)
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
		t, err := compileString(v)
		if err != nil {
			return nil, err
		}
		return t, nil
	case []any:
		return compileChoice(v, pos)
	case map[string]any:
		return compileTemplate(v, pos)
	case nil:
		if pos != inColumn {
			return nil, fmt.Errorf(`null is a record column's value; here it only renders "", so write ""`)
		}
		return &nullItem{}, nil
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

func compileString(s string) (*template, error) {
	toks, unbound, err := parseChecked(s, nil)
	if err != nil {
		return nil, err
	}
	return &template{format: s, tokens: toks, repeat: 1, fromString: true, unbound: unbound}, nil
}

// fixedText is one render's output when the format holds no token; repeat is the caller's.
func (t *template) fixedText() (string, bool) {
	switch {
	case len(t.tokens) == 0:
		return "", true
	case len(t.tokens) == 1 && t.tokens[0].Kind == grammar.LiteralRun:
		return t.tokens[0].Lit, true
	}
	return "", false
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
	c.shared = sharedPaths(c.items, true)
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
		seen[key] = i
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
	if len(fields) == 0 && o.repeat == 1 && !o.weighted && o.datatype == DataTypeString {
		return nil, fmt.Errorf("an object holding only a format is a string; write %q", o.format)
	}
	toks, unbound, err := parseChecked(o.format, fields)
	if err != nil {
		return nil, err
	}
	return &template{format: o.format, tokens: toks, fields: fields, repeat: o.repeat, separator: o.separator, datatype: o.datatype, isRecord: fieldPos == inColumn, unbound: unbound}, nil
}

// templateOptions is what a template object's option keys say.
type templateOptions struct {
	datatype  DataType
	format    string
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

// compileFields compiles every non-option key of a template object.
func compileFields(m map[string]any, pos position) (map[string]node, error) {
	fields := make(map[string]node, len(m))
	for _, k := range sortedNames(m) {
		if isOption(k) {
			continue
		}
		if err := grammar.CheckName(k); err != nil {
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

func isOption(name string) bool {
	switch name {
	case "datatype", "format", "repeat", "separator", "weight":
		return true
	}
	return false
}
