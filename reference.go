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
// refHeads under its root path and its tail reads like a sibling path, so two
// spellings of one target are one draw.
func linkRefs(root map[string]node) error {
	return eachTemplate(root, func(folder []string, path string, t *template) error {
		category := strings.Join(strings.Split(path, ".")[:len(folder)+1], ".")
		return linkTemplate(folder, path, category, t, root)
	})
}

// linkTemplate resolves t's references in category, "" for an inline template, and
// compiles its format against them.
func linkTemplate(folder []string, path, category string, t *template, root map[string]node) error {
	if t.site.table != nil && t.site.row != formatRow {
		path = fmt.Sprintf("%s, line %d", path, t.site.row+2)
	}
	link, err := t.resolveLink(folder, path, category, root)
	if err != nil {
		return err
	}
	compiled, err := compileFormat(t.tokens, link.refs)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	t.link, t.compiled = link, compiled
	return nil
}

// resolveLink binds every reference t reads, refusing one to t's own category, and keys
// t's draw group by that category, so a name is local to it:
// docs/decisions.md#a-category-never-references-itself-and-a-records-fences-run-at-load
func (t *template) resolveLink(folder []string, path, category string, root map[string]node) (templateLink, error) {
	var link templateLink
	if t.drawGroup != "" {
		link.drawGroupKey = category + "/" + t.drawGroup
	}
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
		if err := checkPath(target, tail, head); err != nil {
			return link, fmt.Errorf("%s: reference {%s}: %w", path, name, err)
		}
		link.refHeads[head] = target
		link.refs[name] = refBinding{head, tail}
	}
	link.readsColumn = columnReadOf(t, link)
	return link, nil
}

// columnRead is a record's column read by a format of that one reference alone, which is the
// column: it takes the column's datatype and null.
type columnRead struct {
	a      arm
	column node
}

func columnReadOf(t *template, link templateLink) *columnRead {
	name, lone := loneRef(t.tokens)
	if !lone || t.repeat != 1 {
		return nil
	}
	a := splitArm(name, link.refs)
	target, isTemplate := link.refHeads[a.head].(*template)
	if !isTemplate || !target.isRecord || len(a.tail) != 1 {
		return nil
	}
	return &columnRead{a: a, column: target.fields[a.tail[0]]}
}

// loneRef is the reference a format of one reference token and nothing else reads.
func loneRef(toks []formatToken) (string, bool) {
	if len(toks) != 1 || toks[0].kind != nameRead || len(toks[0].names) != 1 || !isRef(toks[0].names[0]) {
		return "", false
	}
	return toks[0].names[0], true
}

// eachTemplate calls fn once per template, with the folder its category sits in
// and the dot path reaching it, folders and names in sorted order.
func eachTemplate(root map[string]node, fn func(folder []string, path string, t *template) error) error {
	var inCategory func(folder []string, path string, n node) error
	inCategory = func(folder []string, path string, n node) error {
		if t, ok := n.(*template); ok {
			if err := fn(folder, path, t); err != nil {
				return err
			}
		}
		for _, c := range contained(n) {
			if err := inCategory(folder, join(path, c.name), c.node); err != nil {
				return err
			}
		}
		return nil
	}
	var inFolder func(dir []string, children map[string]node) error
	inFolder = func(dir []string, children map[string]node) error {
		for _, name := range sortedNames(children) {
			path := join(strings.Join(dir, "."), name)
			if g, ok := children[name].(*folder); ok {
				if err := inFolder(append(dir[:len(dir):len(dir)], name), g.children); err != nil {
					return err
				}
				continue
			}
			if err := inCategory(dir, path, children[name]); err != nil {
				return err
			}
		}
		return nil
	}
	return inFolder(nil, root)
}

// resolveCategory walks a dotted path through the folders to the category it
// names, returning that category's segments, the node, and the tail left to read into it.
func resolveCategory(root map[string]node, segments []string) (categorySegs []string, target node, tail []string, err error) {
	var n node = &folder{children: root}
	i := 0
	for ; i < len(segments); i++ {
		g, ok := n.(*folder)
		if !ok {
			break
		}
		if isSelector(segments[i]) {
			return nil, nil, nil, fmt.Errorf("%s is a folder, not a table, so it has no row to select", strings.Join(segments[:i], "."))
		}
		child, ok := g.children[segments[i]]
		if !ok {
			return nil, nil, nil, fmt.Errorf("no entry %q", segments[i])
		}
		n = child
	}
	if _, ok := n.(*folder); ok {
		return nil, nil, nil, fmt.Errorf("names a folder, not a value")
	}
	return segments[:i], n, segments[i:], nil
}

// refTokens returns the reference names a format reads, as tokens or as operands.
func refTokens(toks []formatToken) []string {
	var refs []string
	seen := map[string]bool{}
	for _, tok := range toks {
		for _, name := range tok.names {
			if isRef(name) && !seen[name] {
				seen[name] = true
				refs = append(refs, name)
			}
		}
	}
	return refs
}
