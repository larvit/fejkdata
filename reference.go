package fejkdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
)

// refBinding is what a reference resolves to: the head its category is held
// under, and the tail read into it.
type refBinding struct {
	head string
	tail []string
}

// refSegments resolves a reference written in folder to a path from the root.
func refSegments(name string, folder []string) ([]string, error) {
	sigil, rest, err := grammar.RefShape(name)
	if err != nil {
		return nil, err
	}
	var base []string
	switch sigil {
	case ".":
		base = folder
	case "..":
		if len(folder) == 0 {
			return nil, fmt.Errorf("no folder above the root")
		}
		base = folder[:len(folder)-1]
	}
	segs, _ := grammar.SplitPath(rest) // grammar.RefShape proved it splits
	return append(append([]string{}, base...), segs...), nil
}

// linkRefs links every template of the categories, once all data is merged, so a reference sees
// the override-resolved tree.
func linkRefs(sites []categorySite, root map[string]node) error {
	var ts []linkSite
	if err := eachTemplate(sites, func(s categorySite, label string, t *template) error {
		ts = append(ts, linkSite{t: t, folder: s.dir, category: s.path, label: label})
		return nil
	}); err != nil {
		return err
	}
	return linkTemplates(ts, root)
}

// linkSite is a template to link, with the folder and category it sits in, "" for an inline
// template, and eachNode's label for it.
type linkSite struct {
	t               *template
	folder          []string
	category, label string
}

// nameTarget is what a binding resolves to in the assembled tree: the node its head names and the
// path it reads into that node.
type nameTarget struct {
	head node
	tail []string
}

// linkTemplates links ts in steps, each over every template before the next starts, each
// returning what it builds: the references, the names' targets, the compiled formats, the keys
// each name's reads address, and the column each format reads.
func linkTemplates(ts []linkSite, root map[string]node) error {
	for _, s := range ts {
		link, err := s.t.resolveLink(s.folder, s.label, s.category, root)
		if err != nil {
			return err
		}
		s.t.link = link
	}
	targets := nameTargets(ts)
	for _, s := range ts {
		compiled, err := compileFormat(s.t, targets)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		s.t.compiled = compiled
	}
	addressed := addressedKeys(ts)
	for b, target := range targets {
		b.head, b.tail, b.addressed = target.head, target.tail, addressed[b]
	}
	for _, s := range ts {
		s.t.readsColumn = columnReadOf(s.t)
	}
	for _, check := range []func(label string, t *template) error{checkNameReads, checkCalcNames} {
		for _, s := range ts {
			if err := check(s.label, s.t); err != nil {
				return err
			}
		}
	}
	return nil
}

// nameTargets resolves what each binding of ts binds, from its binder's link.
func nameTargets(ts []linkSite) map[*nameBinding]nameTarget {
	targets := map[*nameBinding]nameTarget{}
	for _, s := range ts {
		for _, tok := range s.t.tokens {
			if tok.Kind == grammar.NameBind {
				a := splitArm(tok.BoundRef, s.t.link.refs)
				targets[s.t.nameScope.bindings[tok.Bound]] = nameTarget{head: s.t.head(a.head), tail: a.tail}
			}
		}
	}
	return targets
}

// addressedKeys is every key the reads of each name in ts land on or pass, from the name, with the
// spelling of the first read reaching it.
func addressedKeys(ts []linkSite) map[*nameBinding]map[string]string {
	keys := map[*nameBinding]map[string]string{}
	for _, s := range ts {
		for _, o := range s.t.compiled.ops {
			for _, a := range slices.Concat(o.arms, o.operands) {
				if a.kind != namedRead {
					continue
				}
				if keys[a.named] == nil {
					keys[a.named] = map[string]string{}
				}
				for _, key := range append(a.levels[len(a.levels)-len(a.tail)-1:], a.path) {
					if _, seen := keys[a.named][key]; !seen {
						keys[a.named][key] = a.spelling
					}
				}
			}
		}
	}
	return keys
}

// resolveLink binds every reference t reads, refusing one to t's own category:
// docs/decisions.md#a-category-never-references-itself-and-a-records-fences-run-at-load
func (t *template) resolveLink(folder []string, label, category string, root map[string]node) (templateLink, error) {
	var link templateLink
	names := refTokens(t.tokens)
	if len(names) == 0 {
		return link, nil
	}
	link.refs = make(map[string]refBinding, len(names))
	link.refHeads = make(map[string]node, len(names))
	for _, name := range names {
		segments, err := refSegments(name, folder)
		if err != nil {
			return link, fmt.Errorf("%s: reference {%s}: %w", label, name, err)
		}
		categorySegs, target, tail, err := resolveCategory(root, segments)
		if err != nil {
			return link, fmt.Errorf("%s: reference {%s}: %w", label, name, err)
		}
		head := "/" + strings.Join(categorySegs, ".")
		if category != "" && head == "/"+category {
			return link, fmt.Errorf("%s: reference {%s}: names the category it sits in; read a sibling field as a path, or move the shared value into its own category and reference that", label, name)
		}
		if err := checkPathResolves(target, tail, head); err != nil {
			return link, fmt.Errorf("%s: reference {%s}: %w", label, name, err)
		}
		link.refHeads[head] = target
		link.refs[name] = refBinding{head, tail}
	}
	return link, nil
}

// columnRead is a record's column read by a format that only reads one reference or name,
// which is the column: it takes the column's datatype and null. category and field name it;
// category is "" for a column of the reading template's own record, which checkColumns reaches
// on its own.
type columnRead struct {
	a               arm
	category, field string
	column          node
}

// columnReadOf is the record's column t's format reads, where the format only reads one reference
// or name, and nil where it does not.
func columnReadOf(t *template) *columnRead {
	ops := t.compiled.ops
	if t.repeat != 1 || len(ops) != 1 || ops[0].Kind != grammar.NameRead || len(ops[0].arms) != 1 {
		return nil
	}
	a := ops[0].arms[0]
	head, category, tail := t.head(a.head), a.head[min(1, len(a.head)):], a.tail
	switch b := a.named; {
	case a.kind == namedRead && b.bindsField():
		head, category = b.binder, ""
		tail = append(append([]string{splitArm(b.ref, nil).head}, b.tail...), a.tail...)
	case a.kind == namedRead:
		ref := b.binder.link.refs[b.ref]
		head, category = b.head, ref.head[1:]
		tail = append(ref.tail[:len(ref.tail):len(ref.tail)], a.tail...)
	case !grammar.IsRef(a.head):
		return nil
	}
	if target, isTemplate := head.(*template); isTemplate && target.isRecord && len(tail) == 1 {
		return &columnRead{a: a, category: category, field: tail[0], column: target.fields[tail[0]]}
	}
	return nil
}

// eachTemplate calls fn once per template of the categories, with the category it sits
// in and eachNode's label for it.
func eachTemplate(sites []categorySite, fn func(s categorySite, label string, t *template) error) error {
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(label string, n node) error {
			if t, isTemplate := n.(*template); isTemplate {
				return fn(s, label, t)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// resolveCategory walks a dotted path through the folders to the category it
// names, returning that category's segments, the node, and the tail left to read into it.
func resolveCategory(root map[string]node, segments []string) (categorySegs []string, target node, tail []string, err error) {
	g, i := folderAt(&folder{children: root}, segments)
	switch {
	case i == len(segments):
		return nil, nil, nil, fmt.Errorf("names a folder, not a value")
	case grammar.IsSelector(segments[i]):
		return nil, nil, nil, fmt.Errorf("%s is a folder, not a table, so it has no row to select", strings.Join(segments[:i], "."))
	}
	n, ok := g.children[segments[i]]
	if !ok {
		return nil, nil, nil, fmt.Errorf("no entry %q", segments[i])
	}
	return segments[:i+1], n, segments[i+1:], nil
}

// folderAt is the deepest folder a path from root walks into, and the index of the first
// segment naming no folder in it, len(segs) where every segment does.
func folderAt(root *folder, segs []string) (*folder, int) {
	g := root
	for i, seg := range segs {
		sub, isFolder := g.children[seg].(*folder)
		if !isFolder {
			return g, i
		}
		g = sub
	}
	return g, len(segs)
}

// refTokens returns the reference names a format reads, as tokens or as operands, or binds.
func refTokens(toks []grammar.Token) []string {
	var refs []string
	seen := map[string]bool{}
	for _, tok := range toks {
		names := tokenReads(tok)
		if tok.Kind == grammar.NameBind {
			names = []string{tok.BoundRef}
		}
		for _, name := range names {
			if grammar.IsRef(name) && !seen[name] {
				seen[name] = true
				refs = append(refs, name)
			}
		}
	}
	return refs
}
