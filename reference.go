package fejkdata

import (
	"errors"
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
)

// resolvedRef is what a reference resolves to. head is "/" and the category's path from the root,
// the key the category is held under; tail is the path read into the category.
type resolvedRef struct {
	head string
	tail []string
}

func categoryOf(head string) string { return strings.TrimPrefix(head, "/") }

// refSegments resolves a reference written in folder to a path from the root.
func refSegments(spelling string, folder []string) ([]string, error) {
	sigil, rest, err := grammar.RefShape(spelling)
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

// resolveRefs resolves every reference t reads.
func resolveRefs(t *template, folder []string, label string, root map[string]node) (templateRefs, error) {
	var refs templateRefs
	spellings := refSpellings(t.tokens)
	if len(spellings) == 0 {
		return refs, nil
	}
	refs.bySpelling = make(map[string]resolvedRef, len(spellings))
	refs.categories = make(map[string]node, len(spellings))
	for _, spelling := range spellings {
		segments, err := refSegments(spelling, folder)
		if err != nil {
			return refs, fmt.Errorf("%s: reference {%s}: %w", label, spelling, err)
		}
		categorySegs, target, tail, err := resolveCategory(root, segments)
		if err != nil {
			return refs, fmt.Errorf("%s: reference {%s}: %w", label, spelling, err)
		}
		head := "/" + strings.Join(categorySegs, ".")
		if err := provePath(target, tail, head, 0); err != nil {
			return refs, fmt.Errorf("%s: reference {%s}: %w", label, spelling, err)
		}
		refs.categories[head] = target
		refs.bySpelling[spelling] = resolvedRef{head, tail}
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
		return nil, nil, nil, noEntry{segments[i]}
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

// refSpellings returns the reference spellings a format reads, as tokens or as operands, or binds.
func refSpellings(toks []grammar.Token) []string {
	var refs []string
	seen := map[string]bool{}
	add := func(spelling string) {
		if grammar.IsRef(spelling) && !seen[spelling] {
			seen[spelling] = true
			refs = append(refs, spelling)
		}
	}
	for _, tok := range toks {
		own, clause := tokenReads(tok)
		if tok.Kind == grammar.NameBind {
			own, clause = []string{tok.BoundRef}, nil
		}
		for _, spelling := range own {
			add(spelling)
		}
		for _, c := range clause {
			add(c.spelling)
		}
	}
	return refs
}

// noEntry is a path segment naming nothing in the folder it reads.
type noEntry struct{ seg string }

func (e noEntry) Error() string { return fmt.Sprintf("no entry %q", e.seg) }

// noDataNote adds to err, where a segment names nothing and f holds no data at all, that no
// data is loaded.
func (f *Generator) noDataNote(err error) error {
	var ne noEntry
	if errors.As(err, &ne) && len(f.root.children) == 0 && len(f.root.unloaded) == 0 {
		return fmt.Errorf("%w; no data is loaded", err)
	}
	return err
}
