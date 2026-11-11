package load

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

func categoryOf(head string) string { return strings.TrimPrefix(head, "/") }

// RefSegments resolves a reference written in folder to a path from the root.
func RefSegments(spelling string, folder []string) ([]string, error) {
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
func resolveRefs(t *tree.Template, folder []string, label string, root map[string]tree.Node) (tree.TemplateRefs, error) {
	spellings := RefSpellings(t.Tokens())
	if len(spellings) == 0 {
		return tree.TemplateRefs{}, nil
	}
	bySpelling, categories := make(map[string]tree.ResolvedRef, len(spellings)), make(map[string]tree.Node, len(spellings))
	for _, spelling := range spellings {
		segments, err := RefSegments(spelling, folder)
		if err != nil {
			return tree.TemplateRefs{}, fmt.Errorf("%s: reference {%s}: %w", label, spelling, err)
		}
		categorySegs, target, tail, err := ResolveCategory(root, segments)
		if err != nil {
			return tree.TemplateRefs{}, fmt.Errorf("%s: reference {%s}: %w", label, spelling, err)
		}
		head := "/" + strings.Join(categorySegs, ".")
		if err := provePath(target, tail, head, 0); err != nil {
			return tree.TemplateRefs{}, fmt.Errorf("%s: reference {%s}: %w", label, spelling, err)
		}
		categories[head] = target
		bySpelling[spelling] = tree.NewResolvedRef(head, tail)
	}
	return tree.NewTemplateRefs(bySpelling, categories), nil
}

// ResolveCategory walks a dotted path through the folders to the category it
// names, returning that category's segments, the node, and the tail left to read into it.
func ResolveCategory(root map[string]tree.Node, segments []string) (categorySegs []string, target tree.Node, tail []string, err error) {
	g, i := FolderAt(tree.NewFolderOf(root), segments)
	switch {
	case i == len(segments):
		return nil, nil, nil, fmt.Errorf("names a folder, not a value")
	case grammar.IsSelector(segments[i]):
		return nil, nil, nil, fmt.Errorf("%s is a folder, not a table, so it has no row to select", strings.Join(segments[:i], "."))
	}
	n, ok := g.Children()[segments[i]]
	if !ok {
		return nil, nil, nil, NoEntry{segments[i]}
	}
	return segments[:i+1], n, segments[i+1:], nil
}

// FolderAt is the deepest folder a path from root walks into, and the index of the first
// segment naming no folder in it, len(segs) where every segment does.
func FolderAt(root *tree.Folder, segs []string) (*tree.Folder, int) {
	g := root
	for i, seg := range segs {
		sub, isFolder := g.Children()[seg].(*tree.Folder)
		if !isFolder {
			return g, i
		}
		g = sub
	}
	return g, len(segs)
}

// RefSpellings returns the reference spellings a format reads, as tokens or as operands, or binds.
func RefSpellings(toks []grammar.Token) []string {
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

// NoEntry is a path segment naming nothing in the folder it reads.
type NoEntry struct{ seg string }

func (e NoEntry) Error() string { return fmt.Sprintf("no entry %q", e.seg) }
