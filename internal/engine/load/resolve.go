package load

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// TemplateSite is a template to resolve, with the folders holding its category and the label
// its errors carry.
type TemplateSite struct {
	Template *tree.Template
	Folder   []string
	Label    string
}

// ResolveInlineTemplates resolves every template in n, labelled from label, against the categories
// in root.
func ResolveInlineTemplates(n tree.Node, label string, root map[string]tree.Node, l *tree.LoadFinds) ([]TemplateSite, error) {
	var ts []TemplateSite
	if err := EachNode(n, label, func(label string, n tree.Node) error {
		if t, isTemplate := n.(*tree.Template); isTemplate {
			ts = append(ts, TemplateSite{Template: t, Label: label})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := CompileTemplates(ts, root, l); err != nil {
		return nil, err
	}
	return ts, nil
}

// IndexPaths records in index, for n and each node under it, the segments reaching it from
// where path reaches n; a node already in index keeps its first path.
func IndexPaths(n tree.Node, path []string, index map[tree.Node][]string) {
	if _, seen := index[n]; seen {
		return
	}
	index[n] = path
	for _, c := range Contained(n) {
		below := path
		if c.name != "" {
			below = append(path[:len(path):len(path)], c.name)
		}
		IndexPaths(c.Node, below, index)
	}
	if t, isTable := n.(*tree.Table); isTable {
		for _, linked := range t.ChildTables() {
			IndexPaths(linked, append(path[:len(path):len(path)], linked.Rows().Segment()), index)
		}
	}
}

func compileFormats(ts []TemplateSite, l *tree.LoadFinds) error {
	for _, s := range ts {
		compiled, err := compileFormat(s.Template, l)
		if err != nil {
			return fmt.Errorf("%s: %w", s.Label, err)
		}
		l.SetCompiled(s.Template, compiled)
	}
	return nil
}

// CompileTemplates compiles ts into l in stages, each reading what the stages before it wrote:
//
//  1. resolveAllRefs resolves each template's references.
//  2. setNameTargets finds what each name binds, from its binder's references.
//  3. compileFormats compiles each format against its references and its names' targets.
//  4. buildOps gives each op its call and draw. It needs every format compiled, since an adddays
//     call takes the layout its date operand prints in.
//  5. setColumnReads and setAddressed find the column each template reads, and what each name's
//     reads address, from the finished ops.
func CompileTemplates(ts []TemplateSite, root map[string]tree.Node, l *tree.LoadFinds) error {
	if err := resolveAllRefs(ts, root, l); err != nil {
		return err
	}
	setNameTargets(ts, l)
	if err := compileFormats(ts, l); err != nil {
		return err
	}
	if err := buildOps(ts, l); err != nil {
		return err
	}
	setColumnReads(ts, l)
	setAddressed(ts, l)
	return nil
}

func resolveAllRefs(ts []TemplateSite, root map[string]tree.Node, l *tree.LoadFinds) error {
	for _, s := range ts {
		refs, err := resolveRefs(s.Template, s.Folder, s.Label, root)
		if err != nil {
			return err
		}
		l.SetRefs(s.Template, refs)
	}
	return nil
}

func setNameTargets(ts []TemplateSite, l *tree.LoadFinds) {
	for b, target := range nameTargets(ts, l) {
		l.SetTarget(b, target)
	}
}

func setColumnReads(ts []TemplateSite, l *tree.LoadFinds) {
	for _, s := range ts {
		l.SetReadsColumn(s.Template, columnReadOf(s.Template, l))
	}
}

func setAddressed(ts []TemplateSite, l *tree.LoadFinds) {
	for b, keys := range tree.AddressedKeys(ReadsOfNames(ts, l), l) {
		l.SetAddressed(b, keys)
	}
}

// buildOps completes each op of ts, once every template's format is compiled, and gives l the draw
// of each name bound to a token with a where or a distinct clause.
func buildOps(ts []TemplateSite, l *tree.LoadFinds) error {
	words := map[tree.WordsAt]*tree.WordList{}
	for _, s := range ts {
		tl := l.Of(s.Template)
		for i, parts := range tl.Compiled().Ops() {
			o, err := newOp(s.Template, parts, l, words)
			if err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.Label, parts.Token().Body, err)
			}
			l.SetOp(s.Template, i, o)
			if v, ok := o.Draw().(*tree.ValueDraw); ok && o.Token().Kind == grammar.NameBind {
				l.SetValues(tl.NameScope().Bindings()[o.Token().Bound], v)
			}
		}
	}
	return nil
}
