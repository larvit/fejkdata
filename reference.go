package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
)

// resolvedRef is what a reference resolves to: the head its category is held
// under, and the tail read into it.
type resolvedRef struct {
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

// resolveRefs resolves every reference t reads, refusing one to t's own category:
// docs/decisions.md#a-category-never-references-itself-and-a-records-fences-run-at-load
func (t *template) resolveRefs(folder []string, label, category string, root map[string]node) (templateRefs, error) {
	var refs templateRefs
	names := refTokens(t.tokens)
	if len(names) == 0 {
		return refs, nil
	}
	refs.byName = make(map[string]resolvedRef, len(names))
	refs.categories = make(map[string]node, len(names))
	for _, name := range names {
		segments, err := refSegments(name, folder)
		if err != nil {
			return refs, fmt.Errorf("%s: reference {%s}: %w", label, name, err)
		}
		categorySegs, target, tail, err := resolveCategory(root, segments)
		if err != nil {
			return refs, fmt.Errorf("%s: reference {%s}: %w", label, name, err)
		}
		head := "/" + strings.Join(categorySegs, ".")
		if category != "" && head == "/"+category {
			return refs, fmt.Errorf("%s: reference {%s}: names the category it sits in; read a sibling field as a path, or move the shared value into its own category and reference that", label, name)
		}
		if err := checkPathReaches(target, tail, head); err != nil {
			return refs, fmt.Errorf("%s: reference {%s}: %w", label, name, err)
		}
		refs.categories[head] = target
		refs.byName[name] = resolvedRef{head, tail}
	}
	return refs, nil
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
