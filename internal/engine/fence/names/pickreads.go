package names

import (
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// pickReads is the reads of one name: its binding, the reads through it, and the keys they
// address, each with the spelling of the first read reaching it.
//
// The pick's keys start at the start of the binding's target, and the binding itself lands a
// level for each segment of the target's tail below that. So a read through the name passes those
// levels before the first one it addresses, and a key under the binding opens with the tail. Only
// this file counts the tail's segments: a rule asks the pick.
type pickReads struct {
	binding   *tree.NameBinding
	reads     []*tree.Read
	addressed map[tree.PickKey]string
}

func (names resolvedNames) pick(b *tree.NameBinding) pickReads {
	return pickReads{binding: b, reads: names.reads[b], addressed: b.Addressed()}
}

// addresser is the spelling of the first read of the name that lands on or passes key.
func (p pickReads) addresser(key tree.PickKey) (string, bool) {
	by, addressed := p.addressed[key]
	return by, addressed
}

// landingOn is the spelling of a read that lands on key.
func (p pickReads) landingOn(key tree.PickKey) (string, bool) {
	for _, r := range p.reads {
		if r.Key() == key {
			return r.Spelling(), true
		}
	}
	return "", false
}

// addresses reports whether r, a read of the name, lands on or passes key.
func (p pickReads) addresses(r *tree.Read, key tree.PickKey) bool {
	return slices.Contains(r.Levels()[len(p.binding.Target().Tail()):], key)
}

// keyAt is the key of the node at path, the fields leading down to it from the start of the
// binding's target. A key holds the selector of a binding such as {/region[01] as r}, and a walk
// of the tree passes no selector.
func (p pickReads) keyAt(path []string) tree.PickKey {
	tail := p.binding.Target().Tail()
	if len(tail) > 0 && grammar.IsSelector(tail[0]) {
		return keyOf(tail[:1], path)
	}
	return keyOf(path)
}

// spellingOf is how to read level, a key in the pick, through the name.
func (p pickReads) spellingOf(level tree.PickKey) string {
	b := p.binding
	bound := len(b.Target().Tail())
	segs, err := grammar.SplitPath(string(level))
	if err != nil || len(segs) < bound {
		panic(invariant.Broken("level %q of name %q is not a path under its binding", level, b.Name()))
	}
	return grammar.JoinSegments(append([]string{b.Name()}, segs[bound:]...))
}

// keyOf joins the path segments in parts, in order, into the key a pick keeps that level under.
func keyOf(parts ...[]string) tree.PickKey {
	return tree.PickKey(strings.Join(slices.Concat(parts...), "."))
}
