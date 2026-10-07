package fejkdata

import (
	"fmt"
	"math"

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

// choice draws one of its items. cum holds cumulative weights for a weighted
// draw; when nil the choice is uniform and selection is O(1). shared is the set of
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
// an independent draw. Every format compiles in `resolveTemplates`.
type template struct {
	// Filled by `compileString`, `compileTemplate` and `table.compileRowFormat`:
	format     string
	tokens     []grammar.Token
	fields     map[string]node
	repeat     int
	separator  string
	datatype   DataType
	declared   bool          // carries a "datatype", which wins over the datatype of a column it reads
	fromString bool          // written as a JSON string rather than an object
	isRecord   bool          // compiled at the top without a repeat, so its fields are record columns
	unbound    []unboundRead // the heads its tokens read that no field holds

	// Filled by `bindNames`, from the compiled category:
	nameScope *nameScope // where its tokens look a name up

	// Filled by `resolveTemplates`, from the assembled tree:
	refs        templateRefs
	compiled    formatOps
	readsColumn *columnRead // set when the format only reads one reference or name, and that read is a record's column

	// Filled by `settleRecords`, once the load refused every cycle:
	columns []recordColumn // a record's, where it has fields, in name order
}

// templateRefs is what a template resolves to in the assembled tree.
type templateRefs struct {
	byName     map[string]resolvedRef // each reference the format reads -> what it resolves to
	categories map[string]node        // each resolvedRef.head -> the category it names
}

func (*template) isNode() {}

// startOf is the node an arm's head names: a sibling field, or a reference's category.
func (t *template) startOf(name string) node {
	if grammar.IsRef(name) {
		return t.refs.categories[name]
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
// file beside it.
func compileCategory(c datafiles.Category) (node, error) {
	if m, ok := c.JSON.(map[string]any); ok {
		if _, isTable := m["rows"]; isTable {
			return compileTable(m, c.Folders, c.Name, c.ReadRows)
		}
	}
	return compile(c.JSON)
}

// position is where a JSON value sits, which decides whether it may carry a datatype or
// be null.
type position int

const (
	inFormat position = iota // rendered by a format, so neither
	atTop                    // a category or an inline template, whose fields may be columns
	inColumn                 // a column, or a choice item standing in for one
)

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
			return compileString("")
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
	switch len(items) {
	case 0:
		return nil, fmt.Errorf("empty choice")
	case 1:
		return compileAt(items[0], pos)
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
	if weighted { // uniform choices skip the weight table and draw in O(1)
		if math.IsInf(total, 1) {
			return nil, fmt.Errorf("choice weights must sum to a finite number, got %v", total)
		}
		c.cum = cum
	}
	c.shared = sharedPaths(c.items, true)
	return c, nil
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
	toks, unbound, err := parseChecked(o.format, fields)
	if err != nil {
		return nil, err
	}
	return &template{format: o.format, tokens: toks, fields: fields, repeat: o.repeat, separator: o.separator, datatype: o.datatype, declared: o.declared, isRecord: fieldPos == inColumn, unbound: unbound}, nil
}

// templateOptions is what a template object's option keys say.
type templateOptions struct {
	datatype  DataType
	declared  bool
	format    string
	repeat    int
	separator string
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
	_, o.declared = m["datatype"]
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
	}
	return o, nil
}

func compileFields(m map[string]any, pos position) (map[string]node, error) {
	fields := make(map[string]node, len(m))
	for _, k := range sortedNames(m) {
		if isOption(k) {
			continue
		}
		if err := grammar.CheckIdentifier(k); err != nil {
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
// is rendered and concatenated. A present one must be a positive integer.
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
	return w, nil
}

func isOption(name string) bool {
	switch name {
	case "datatype", "format", "repeat", "separator", "weight":
		return true
	}
	return false
}
