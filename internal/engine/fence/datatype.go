package fence

import (
	"errors"
	"fmt"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// CheckColumns rejects a record column whose items hold different datatypes, checking a column
// after the other categories' columns its items read, so a column read is named before its readers.
func CheckColumns(s NodeSet) error {
	checked := map[tree.Node]bool{}
	var check func(label, name string, column tree.Node) error
	check = func(label, name string, column tree.Node) error {
		if checked[column] {
			return nil
		}
		checked[column] = true
		items, _ := load.ColumnItems(column)
		if err := checkReadColumns(label, name, items, check); err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		first := load.ColumnDatatype(column, tree.AfterFinish)
		for _, it := range items[1:] {
			if d := load.ItemDatatype(it, tree.AfterFinish); d != first {
				return fmt.Errorf("%s: field %q: %w", label, name, disagreement(items[0], first, it, d))
			}
		}
		return nil
	}
	return s(func(label string, n tree.Node) error {
		t, ok := n.(*tree.Template)
		if !ok || !t.IsRecord() {
			return nil
		}
		for _, name := range tree.SortedNames(t.Fields()) {
			if err := check(label, name, t.Fields()[name]); err != nil {
				return err
			}
		}
		return nil
	})
}

// itemKind is how a column item comes by its datatype, ranked in this order: none, read from a typed column, or declared.
type itemKind int

const (
	kindText itemKind = iota
	kindReads
	kindDeclares
)

func kindOf(t *tree.Template) itemKind {
	switch {
	case t.DataType() != nil:
		return kindDeclares
	case load.ItemDatatype(t, tree.AfterFinish) != datatype.String:
		return kindReads
	}
	return kindText
}

// disagreement names the fix for items a and b of one column holding different datatypes, fixing
// the one that declares least, a when both declare alike.
func disagreement(a *tree.Template, da datatype.DataType, b *tree.Template, db datatype.DataType) error {
	c := clash{fix: a, other: b, want: db, held: fmt.Sprintf("its items hold %s and %s; a column holds one datatype", da, db)}
	if kindOf(b) < kindOf(a) {
		c.fix, c.other, c.want = b, a, da
	}
	message, ok := disagreements[[2]itemKind{kindOf(c.fix), kindOf(c.other)}]
	if !ok {
		panic(invariant.Broken("disagreement has no message for item kinds %d and %d; two text items hold one datatype, so they never disagree", kindOf(c.fix), kindOf(c.other)))
	}
	return message(c)
}

// clash is two items of one column holding different datatypes: other holds want; held says
// what each holds.
type clash struct {
	fix, other *tree.Template
	want       datatype.DataType
	held       string
}

// disagreements is a clash's message by the kinds of its items, fix's first.
var disagreements = map[[2]itemKind]func(clash) error{
	{kindDeclares, kindDeclares}: clash.bothDeclare,
	{kindReads, kindReads}:       clash.retypeRead,
	{kindReads, kindDeclares}:    clash.retypeRead,
	{kindText, kindReads}:        clash.textBesideRead,
	{kindText, kindDeclares}:     clash.declareText,
}

func (c clash) bothDeclare() error { return errors.New(c.held) }

// fits reports whether fix may hold want.
func (c clash) fits() bool { return (&valueProof{}).proveColumnItem(c.fix).Not[c.want] == "" }

func (c clash) retypeRead() error {
	if c.fits() {
		return fmt.Errorf("%s, so %s", c.held, typedAs(c.fix, c.want))
	}
	return fmt.Errorf("%s, so to read %q as text, %s", c.held, c.fix.Format(), typedAs(c.fix, datatype.String))
}

func (c clash) textBesideRead() error {
	if !c.fits() {
		return fmt.Errorf(`item %q is not %s, the datatype item %q takes from the column it reads; to read that column as text, %s`, c.fix.Format(), datatype.Noun(c.want), c.other.Format(), typedAs(c.other, datatype.String))
	}
	return c.declareText()
}

func (c clash) declareText() error {
	if c.fix.FromString() { // an object may carry a weight, which this spelling would drop
		return fmt.Errorf(`item %q declares no datatype, and a column holds one; write it as {"format":%q,"datatype":%q}`, c.fix.Format(), c.fix.Format(), c.want)
	}
	return fmt.Errorf(`item %q declares no datatype beside one holding %s; a column holds one, so give it "datatype": %q`, c.fix.Format(), c.want, c.want)
}

// typedAs names the spelling giving a column-read item datatype d, keeping the other keys an object
// item carries.
func typedAs(t *tree.Template, d datatype.DataType) string {
	if t.FromString() {
		return fmt.Sprintf(`write %q as {"format":%q,"datatype":%q}`, t.Format(), t.Format(), d)
	}
	return fmt.Sprintf(`give %q "datatype": %q`, t.Format(), d)
}
