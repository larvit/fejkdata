package fejkdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/jsonvalue"
	"github.com/larvit/fejkdata/internal/proven"
)

// compile converts parsed JSON — a category or an inline template — into a node tree,
// validating structure up front, and gives its names their scopes in l.
func compile(v jsonvalue.Value, l *tree.LoadFinds) (tree.Node, error) {
	return compileFrom(v, atTop, l)
}

func compileFrom(v jsonvalue.Value, pos position, l *tree.LoadFinds) (tree.Node, error) {
	n, err := compileItem(v.Any(), pos)
	if err != nil {
		return nil, err
	}
	return n, bindNames(n, l)
}

// compileCategory compiles a data file's value, which may be a table over a rows
// file beside it.
func compileCategory(c datafiles.Category, l *tree.LoadFinds) (tree.Node, error) {
	if m, ok := c.JSON.Any().(map[string]any); ok {
		if _, isTable := m["rows"]; isTable {
			return compileTable(m, c.Folders, c.Name, c.ReadRows)
		}
	}
	return compile(c.JSON, l)
}

// position is where a JSON value sits, which decides whether it may carry a datatype or
// be null.
type position int

const (
	inFormat      position = iota // rendered by a format, so neither
	atTop                         // a category or an inline template, whose fields may be columns
	inColumn                      // a column, or a choice item standing in for one
	atGoTop                       // a struct's tags, whose fields are columns a Go type types
	inGoColumn                    // a column a struct field's Go type types, or a choice item in one
	inColumnField                 // a field of a template standing as a column: may be null or typed, a bare number stays text
)

// column reports a position that is a record column.
func (p position) column() bool { return p == inColumn || p == inGoColumn || p == inColumnField }

// compileItem compiles one node; a weight on it skews only a choice drawing it.
func compileItem(v any, pos position) (tree.Node, error) {
	if _, err := weightOf(v); err != nil {
		return nil, err
	}
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
		if !pos.column() {
			return compileString("")
		}
		return &tree.NullItem{}, nil
	case json.Number:
		if pos == inColumn {
			return nil, typedOrText(string(v), DataTypeInteger, DataTypeNumber)
		}
		return compileString(string(v))
	case bool:
		if pos == inColumn {
			return nil, typedOrText(strconv.FormatBool(v), DataTypeBoolean)
		}
		return compileString(strconv.FormatBool(v))
	}
	panic(invariant.Broken("compileItem has no case for the JSON value %T", v))
}

// typedOrText refuses a bare JSON number or boolean in a record column, which could be the
// text it spells or a typed value, naming the text and the first of ds that holds it.
func typedOrText(text string, ds ...DataType) error {
	held := proven.Literal(text)
	for _, d := range ds {
		if held.Not[d] == "" {
			return fmt.Errorf(`%s in a record column could be text or a value of datatype %s; write %q for text, or {"format":%q,"datatype":%q}`, text, d, text, text, d)
		}
	}
	return fmt.Errorf(`%s in a record column fits no datatype; write %q for text`, text, text)
}

func compileString(s string) (*tree.Template, error) {
	toks, unbound, err := parseChecked(s, nil)
	if err != nil {
		return nil, err
	}
	return tree.NewTemplate(tree.TemplateSyntax{Format: s, Tokens: toks, Repeat: 1, FromString: true, Unbound: unbound}), nil
}

// printedOp is the one op of ops that prints, past the binds a where or distinct clause compiles to,
// which print nothing; ok is false where more than one op prints, or none does.
func printedOp(ops []tree.Op) (printed tree.Op, ok bool) {
	n := 0
	for _, o := range ops {
		if o.Token().Kind != grammar.NameBind {
			printed, n = o, n+1
		}
	}
	return printed, n == 1
}

func compileChoice(items []any, pos position) (tree.Node, error) {
	switch len(items) {
	case 0:
		return nil, fmt.Errorf("empty choice")
	case 1:
		return compileLoneItem(items[0], pos)
	}
	itemPos := inFormat
	if pos.column() {
		itemPos = pos
	}
	compiled := make([]tree.Node, len(items))
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
		if compiled[i], err = compileItem(raw, itemPos); err != nil {
			return nil, err
		}
	}
	if !weighted {
		cum = nil
	} else if err := checkWeightSum(total); err != nil {
		return nil, err
	}
	return tree.NewChoice(compiled, cum, sharedPaths(compiled, true)), nil
}

// compileLoneItem compiles a choice of one item, which is that item.
func compileLoneItem(raw any, pos position) (tree.Node, error) {
	w, err := weightOf(raw)
	if err != nil {
		return nil, err
	}
	if w == 0 {
		return nil, errNothingToDraw
	}
	return compileItem(raw, pos)
}

var errNothingToDraw = errors.New("every weight is 0, so the choice has nothing to draw")

func checkWeightSum(total float64) error {
	if math.IsInf(total, 1) {
		return fmt.Errorf("choice weights must sum to a finite number, got %v", total)
	}
	if total == 0 {
		return errNothingToDraw
	}
	return nil
}

func compileTemplate(m map[string]any, pos position) (tree.Node, error) {
	if _, isTable := m["rows"]; isTable {
		return nil, fmt.Errorf("rows names a TSV beside a category's file, so only a category is a table; an inline template has no file beside it")
	}
	o, err := readOptions(m, pos)
	if err != nil {
		return nil, err
	}
	fieldPos := inFormat
	if o.repeat == 1 {
		switch pos {
		case atTop:
			fieldPos = inColumn
		case atGoTop:
			fieldPos = inGoColumn
		case inColumn, inColumnField:
			fieldPos = inColumnField
		}
	}
	fields, lets, err := compileFields(m, fieldPos)
	if err != nil {
		return nil, err
	}
	toks, unbound, err := parseChecked(o.format, fields)
	if err != nil {
		return nil, err
	}
	return tree.NewTemplate(tree.TemplateSyntax{Format: o.format, Tokens: toks, Fields: fields, Lets: lets, Repeat: o.repeat, Separator: o.separator, DataType: o.datatype, IsRecord: fieldPos.column(), Unbound: unbound}), nil
}

type templateOptions struct {
	datatype  *DataType
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
	if o.datatype, err = datatypeOf(m, pos); err != nil {
		return o, err
	}
	if sv, ok := m["separator"]; ok {
		if o.separator, ok = sv.(string); !ok {
			return o, fmt.Errorf("separator must be a string, not %s", jsonvalue.Kind(sv))
		}
	}
	return o, nil
}

// compileFields compiles the fields beside the options, and those the "let" option holds, which it
// also names in lets. A key of let that is an option name fails, as a field cannot take it.
func compileFields(m map[string]any, pos position) (fields map[string]tree.Node, lets map[string]bool, err error) {
	fields = make(map[string]tree.Node, len(m))
	compile := func(raw map[string]any, skipOptions bool) error {
		for _, k := range tree.SortedNames(raw) {
			if isOption(k) {
				if skipOptions {
					continue
				}
				return fmt.Errorf("let holds fields, and %q is an option; name the field otherwise", k)
			}
			if err := grammar.CheckIdentifier(k); err != nil {
				return fmt.Errorf("field %w", err)
			}
			if _, twice := fields[k]; twice {
				return fmt.Errorf("field %q is in let and beside it; write it once", k)
			}
			n, err := compileItem(raw[k], pos)
			if err != nil {
				return fmt.Errorf("field %q: %w", k, err)
			}
			fields[k] = n
		}
		return nil
	}
	if err := compile(m, true); err != nil {
		return nil, nil, err
	}
	let, hasLet := m["let"]
	if !hasLet {
		return fields, nil, nil
	}
	held, ok := let.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("let is an option and takes an object of fields, not %s; a field cannot be named \"let\"", jsonvalue.Kind(let))
	}
	if err := compile(held, false); err != nil {
		return nil, nil, err
	}
	lets = make(map[string]bool, len(held))
	for k := range held {
		lets[k] = true
	}
	return fields, lets, nil
}

// repeatOf reads a template's "repeat" (default 1): how many times its format
// is rendered and concatenated. A present one must be an integer of 0 or more.
func repeatOf(m map[string]any) (int, error) {
	rv, ok := m["repeat"]
	if !ok {
		return 1, nil
	}
	r, ok := numberOf(rv)
	if !ok {
		return 0, fmt.Errorf("repeat must be a number, not %s", jsonvalue.Kind(rv))
	}
	if err := notTiny("repeat", rv, r); err != nil {
		return 0, err
	}
	if math.IsNaN(r) || math.IsInf(r, 0) || r < 0 || r != math.Trunc(r) {
		return 0, fmt.Errorf("repeat must be an integer of 0 or more, got %v", rv)
	}
	if r > MaxRepeat { // caps the renders one repeat asks for; repeatCheck bounds what nested ones multiply to
		return 0, fmt.Errorf("repeat %v exceeds the maximum %d", rv, MaxRepeat)
	}
	return int(r), nil
}

// weightOf reads a node's "weight" (default 1) from its raw JSON form. Only
// template objects carry weight; a present one must be finite and not negative.
func weightOf(raw any) (float64, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return 1, nil
	}
	wv, ok := m["weight"]
	if !ok {
		return 1, nil
	}
	w, ok := numberOf(wv)
	if !ok {
		return 0, fmt.Errorf(`weight is an option and takes a number, not %s; a field cannot be named "weight"`, jsonvalue.Kind(wv))
	}
	if w < 0 || math.IsInf(w, 0) {
		return 0, fmt.Errorf("weight must be a finite number of 0 or more, got %v", wv)
	}
	return w, notTiny("weight", wv, w)
}

// notTiny refuses the number v of option name where grammar.Underflows does.
func notTiny(name string, v any, f float64) error {
	if grammar.Underflows(string(v.(json.Number)), f) {
		return fmt.Errorf("%s %s is too close to 0 to tell from it", name, v)
	}
	return nil
}

// numberOf reads a JSON number, one too large for a float64 as an infinity.
func numberOf(v any) (float64, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	f, _ := n.Float64()
	return f, true
}

func isOption(name string) bool {
	switch name {
	case "datatype", "format", "let", "repeat", "separator", "weight":
		return true
	}
	return false
}
