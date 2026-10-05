package fejkdata

import (
	"fmt"
	"strings"
)

// A reference names a node by path rather than as a sibling field: {/a.b} from the
// data root, {.a} from the folder this file sits in, {..a} from the folder above.
func isRef(name string) bool { return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "/") }

// refBinding is what a reference resolves to: the head its category is held
// under, and the tail read into it.
type refBinding struct {
	head string
	tail []string
}

// refShape splits a reference into its sigil and the dotted path after it.
func refShape(name string) (sigil, rest string, err error) {
	switch {
	case strings.HasPrefix(name, "/"):
		sigil, rest = "/", name[1:]
	case strings.HasPrefix(name, ".."):
		sigil, rest = "..", name[2:]
	default:
		sigil, rest = ".", name[1:]
	}
	if rest == "" {
		return "", "", fmt.Errorf("reference has no path")
	}
	if strings.HasPrefix(rest, "/") {
		return "", "", fmt.Errorf("the path after %s starts at a name, not a /; write {%s%s}", sigil, sigil, rest[1:])
	}
	if strings.HasPrefix(rest, ".") {
		return "", "", fmt.Errorf("a reference starts with / (the root), . (this folder) or .. (the folder above)")
	}
	segs, err := splitPath(rest)
	if err != nil {
		return "", "", err
	}
	for _, seg := range segs {
		if seg == "" {
			return "", "", fmt.Errorf("path has an empty segment")
		}
	}
	return sigil, rest, nil
}

// refSegments resolves a reference written in folder to a path from the root.
func refSegments(name string, folder []string) ([]string, error) {
	sigil, rest, err := refShape(name)
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
	segs, _ := splitPath(rest) // refShape proved it splits
	return append(append([]string{}, base...), segs...), nil
}

// linkRefs binds every reference and compiles every format, once all data is merged,
// so a reference sees the override-resolved tree. A reference's head binds into
// refHeads under its root path and its tail reads like a sibling path.
func linkRefs(sites []categorySite, root map[string]node) error {
	if err := eachTemplate(sites, func(s categorySite, path string, t *template) error {
		return linkTemplate(s.dir, path, s.path, t, root)
	}); err != nil {
		return err
	}
	for _, pass := range linkPasses {
		if err := eachTemplate(sites, func(_ categorySite, path string, t *template) error { return pass(path, t) }); err != nil {
			return err
		}
	}
	return nil
}

// linkTemplate resolves t's references in category, "" for an inline template, and
// compiles its format against them.
func linkTemplate(folder []string, path, category string, t *template, root map[string]node) error {
	link, err := t.resolveLink(folder, t.site.label(path), category, root)
	if err != nil {
		return err
	}
	t.link, t.compiled = link, compileFormat(t.tokens, link.refs, t.isName)
	linkBindings(t)
	compileArms(t)
	return nil
}

func compileArms(t *template) {
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		for j := range o.operands {
			if o.operands[j].kind != namedRead {
				compileArm(t, &o.operands[j])
			}
		}
		for j := range o.arms {
			if o.arms[j].kind != namedRead {
				compileArm(t, &o.arms[j])
			}
		}
	}
}

func compileArm(t *template, a *arm) {
	head := t.head(a.head)
	if head == nil {
		panic(internalError("{%s} reads a head nothing bound", a.spelling))
	}
	w := compilePath(head, a.tail)
	a.steps, a.leaves = w.steps, w.leaves
}

// resolveLink binds every reference t reads, refusing one to t's own category:
// docs/decisions.md#a-category-never-references-itself-and-a-records-fences-run-at-load
func (t *template) resolveLink(folder []string, path, category string, root map[string]node) (templateLink, error) {
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
			return link, fmt.Errorf("%s: reference {%s}: %w", path, name, err)
		}
		categorySegs, target, tail, err := resolveCategory(root, segments)
		if err != nil {
			return link, fmt.Errorf("%s: reference {%s}: %w", path, name, err)
		}
		head := "/" + strings.Join(categorySegs, ".")
		if category != "" && head == "/"+category {
			return link, fmt.Errorf("%s: reference {%s}: names the category it sits in; read a sibling field as a path, or move the shared value into its own category and reference that", path, name)
		}
		if err := checkPathResolves(target, tail, head); err != nil {
			return link, fmt.Errorf("%s: reference {%s}: %w", path, name, err)
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

// linkColumnRead sets the record's column t's format reads, where the format only reads one
// reference or name, once the names are linked.
func linkColumnRead(_ string, t *template) error {
	ops := t.compiled.ops
	if t.repeat != 1 || len(ops) != 1 || ops[0].kind != nameRead || len(ops[0].arms) != 1 {
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
	case !isRef(a.head):
		return nil
	}
	target, isTemplate := head.(*template)
	if isTemplate && target.isRecord && len(tail) == 1 {
		t.link.readsColumn = &columnRead{a: a, category: category, field: tail[0], column: target.fields[tail[0]]}
	}
	return nil
}

// eachTemplate calls fn once per template of the categories, with the category it sits
// in and the dot path reaching it.
func eachTemplate(sites []categorySite, fn func(s categorySite, path string, t *template) error) error {
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(path string, n node) error {
			if t, isTemplate := n.(*template); isTemplate {
				return fn(s, path, t)
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
	case isSelector(segments[i]):
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
func refTokens(toks []formatToken) []string {
	var refs []string
	seen := map[string]bool{}
	for _, tok := range toks {
		names := tok.names
		if tok.kind == nameBind {
			names = []string{tok.boundRef}
		}
		for _, name := range names {
			if isRef(name) && !seen[name] {
				seen[name] = true
				refs = append(refs, name)
			}
		}
	}
	return refs
}
