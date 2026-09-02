package fejkdata

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// node is a compiled template element: a choice or a template. Compiling JSON
// into these once (see compile) means rendering never re-inspects the raw JSON or
// re-sums weights.
type node interface{ isNode() }

// group is a namespace of named children, built from a directory of JSON files
// and subdirectories. It has no value of its own: descend into a named child by
// dot path; rendering one is an error (see Fake).
type group struct{ children map[string]node }

func (*group) isNode() {}

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

// template renders a format string, substituting {tokens} from fields. A bare
// JSON string is a template with no fields. repeat (default 1) renders that format
// that many times and joins the results with separator (default ""), each render
// an independent pick.
type template struct {
	format    string
	fields    map[string]node
	repeat    int
	separator string
	ops       []op                  // format compiled once (see compileOps); what expand walks
	grow      int                   // minimum output size, to size the render buffer
	fixed     bool                  // no op varies, so every render is lit
	lit       string                // the whole output when fixed
	refs      map[string]refBinding // each reference the format reads -> what it is bound to
	// bound maps each field the format addresses by dotted path to one path token
	// reading it, which is the half of an overlap the fences name. nil when the
	// format takes no path.
	bound map[string]string
	// held is every name drawn once per expansion: the bound levels above, plus the
	// siblings a {calc()} reads. nil when the format holds nothing (see expand).
	held map[string]bool
}

func (*template) isNode() {}

// field is the node a path segment names; a binding is a render edge, not a field.
func (t *template) field(seg string) (node, bool) {
	if isRef(seg) {
		return nil, false
	}
	n, ok := t.fields[seg]
	return n, ok
}

// compile converts parsed JSON into a node tree, validating structure up front.
// Only a choice's items carry a weight, so one here would be inert whatever its type.
func compile(v any) (node, error) {
	if m, ok := v.(map[string]any); ok {
		if _, weighted := m["weight"]; weighted {
			return nil, fmt.Errorf("weight only skews a choice's items, so it has no effect here; it is an option and can never be a field")
		}
	}
	return compileItem(v)
}

// compileItem compiles one node, allowing the weight a choice item may carry.
func compileItem(v any) (node, error) {
	switch v := v.(type) {
	case string:
		return compileString(v)
	case []any:
		return compileChoice(v)
	case map[string]any:
		return compileTemplate(v)
	default:
		return nil, fmt.Errorf("unsupported node type %T", v)
	}
}

func compileString(s string) (node, error) {
	if err := checkTokens(s, nil); err != nil {
		return nil, err
	}
	t := &template{format: s, repeat: 1}
	if err := t.compileFormat(); err != nil {
		return nil, err
	}
	return t, nil
}

// compileFormat compiles the format into ops once every field is in place, and
// applies the fences that need the compiled reads.
func (t *template) compileFormat() error {
	c := compileOps(t.format, t.refs)
	t.ops, t.grow, t.bound, t.held = c.ops, c.grow, c.bound, c.held
	t.fixed = true
	for _, o := range t.ops {
		if o.kind != 'l' {
			t.fixed = false
		}
	}
	if t.fixed && len(t.ops) == 1 {
		t.lit = t.ops[0].lit
	}
	if err := checkNoOverlap(t.format, t.bound, t.refs); err != nil {
		return err
	}
	return checkNoRepeatedRead(t.format, c, t.refs)
}

func compileChoice(items []any) (node, error) {
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
		n, err := compileItem(raw)
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
	// Computed before linkRefs binds references into the items: a binding is keyed
	// by a reference sigil, which paths skips, so the set is the same after.
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
			return fmt.Errorf("choice item %d repeats item %d; skew the odds with a weight on one of them instead", i, j)
		}
		seen[string(key)] = i
	}
	return nil
}

func compileTemplate(m map[string]any) (node, error) {
	o, err := readOptions(m)
	if err != nil {
		return nil, err
	}
	fields, err := compileFields(m)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 && o.repeat == 1 && !o.weighted {
		return nil, fmt.Errorf("an object holding only a format is a string; write %q", o.format)
	}
	if err := checkTokens(o.format, fields); err != nil {
		return nil, err
	}
	t := &template{format: o.format, fields: fields, repeat: o.repeat, separator: o.separator}
	if err := t.compileFormat(); err != nil {
		return nil, err
	}
	return t, nil
}

// templateOptions is what a template object's option keys say.
type templateOptions struct {
	format    string
	repeat    int
	separator string
	weighted  bool
}

func readOptions(m map[string]any) (templateOptions, error) {
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
func compileFields(m map[string]any) (map[string]node, error) {
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
		n, err := compile(m[k])
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
	if r > maxLen { // cap so a fat-fingered repeat can't build a multi-GB string
		return 0, fmt.Errorf("repeat %v exceeds the maximum %d", rv, maxLen)
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

// reservedInName is what a category, folder or field name may not contain: a dot
// separates the segments of a path, '|' the arms of a token, '(' opens a function
// call, braces delimit the token and '/' starts a reference. A name carrying one is
// reachable by no format, so it is rejected where it is authored rather than at
// the token that cannot reach it.
const reservedInName = ".|({}/"

// reservedList spells reservedInName for an error message, so the two cannot drift.
var reservedList = strings.Join(strings.Split(reservedInName, ""), " ")

// checkName rejects a name the dot path and {token} grammars cannot spell. Both a
// category or folder and a field go through it, so there is one answer to what a
// name may contain.
func checkName(name string) error {
	if name == "" {
		return fmt.Errorf("%q is empty, which is not a path segment, so List never offers it", name)
	}
	if i := strings.IndexAny(name, reservedInName); i >= 0 {
		return fmt.Errorf("%q contains %q; a name may not use %s, which the dot path and {token} grammars reserve",
			name, name[i:i+1], reservedList)
	}
	return nil
}

// isOption reports whether a template key configures the node instead of naming a
// field. These four names can never be fields.
func isOption(name string) bool {
	switch name {
	case "format", "repeat", "separator", "weight":
		return true
	}
	return false
}
