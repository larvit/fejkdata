package fejkdata

import (
	"fmt"
	"strings"
)

// A reference names a node by path rather than as a sibling field: {/a.b} from the
// data root, {.a} from the folder this file sits in, {..a} from the folder above.
func isRef(name string) bool { return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "/") }

// refBinding is what a reference was bound to: the head key its category is held
// under, and the tail read into it.
type refBinding struct {
	key  string
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

// linkRefs resolves every reference in the assembled tree. The head of the path —
// up to the category it names — is bound into the referring template's refHeads
// under its root path, and the rest reads into it the way a sibling path does, so
// a reference is held like a sibling and two spellings of one target are one
// draw. It runs once, after all data is merged, so a reference sees the final
// (override-resolved) tree. A path that is unknown, names a folder, or reads a
// field not every variant carries fails here, keeping a bad reference a New-time
// error, never a random render-time one.
func linkRefs(root map[string]node) error {
	return eachTemplate(root, func(folder []string, path string, t *template) error {
		category := strings.Join(strings.Split(path, ".")[:len(folder)+1], ".")
		t.keyDrawGroup(category)
		return linkTemplateRefs(folder, path, category, t, root)
	})
}

// linkTemplateRefs binds one template's references against root, refusing one that names the
// category it sits in.
// docs/decisions.md#a-category-never-references-itself-and-a-records-fences-run-at-load
func linkTemplateRefs(folder []string, path, category string, t *template, root map[string]node) error {
	names := refTokens(t.tokens)
	if len(names) == 0 {
		return nil
	}
	if t.cellOf != nil {
		path = fmt.Sprintf("%s, line %d", path, t.cellRow+2)
	}
	t.refs = make(map[string]refBinding, len(names))
	t.refHeads = make(map[string]node, len(names))
	for _, name := range names {
		segments, err := refSegments(name, folder)
		if err != nil {
			return fmt.Errorf("%s: reference {%s}: %w", path, name, err)
		}
		head, target, tail, err := resolveCategory(root, segments)
		if err != nil {
			return fmt.Errorf("%s: reference {%s}: %w", path, name, err)
		}
		key := "/" + strings.Join(head, ".")
		if category != "" && key == "/"+category {
			return fmt.Errorf("%s: reference {%s}: names the category it sits in; read a sibling field as a path, or move the shared value into its own category and reference that", path, name)
		}
		if err := checkPath(target, tail, key); err != nil {
			return fmt.Errorf("%s: reference {%s}: %w", path, name, err)
		}
		t.refHeads[key] = target
		t.refs[name] = refBinding{key, tail}
	}
	if err := t.compileFormat(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	t.readsColumn = columnReadOf(t)
	return nil
}

// columnRead is a record's column read by a format of that one reference alone, which is the
// column: it takes the column's datatype and null.
type columnRead struct {
	a      arm
	column node
}

func columnReadOf(t *template) *columnRead {
	name, lone := loneRef(t.tokens)
	if !lone || t.repeat != 1 {
		return nil
	}
	a := splitArm(name, t.refs)
	target, isTemplate := t.head(a.key).(*template)
	if !isTemplate || !target.isRecord || len(a.tail) != 1 {
		return nil
	}
	return &columnRead{a: a, column: target.fields[a.tail[0]]}
}

// loneRef is the reference a format of one reference token and nothing else reads.
func loneRef(toks []formatToken) (string, bool) {
	if len(toks) != 1 || toks[0].kind != 'f' || len(toks[0].names) != 1 || !isRef(toks[0].names[0]) {
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
// names, returning that head, the node, and the tail left to read into it. A
// descent of its own rather than a walkPath: it walks folders only and returns
// where they end, not a leaf.
func resolveCategory(root map[string]node, segments []string) (head []string, target node, tail []string, err error) {
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
