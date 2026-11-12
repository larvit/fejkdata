package tree

import "github.com/larvit/fejkdata/internal/grammar"

// PickKey is the key a named pick's memo keeps one level's draw under.
//
// A level is one prefix of a read's path, and a read has one for each node it passes; a choice
// and the variant drawn from it, or a table and its row, share one. A level's key is its path from
// where the read starts, and LevelsUnder says where that is. A load gives each read one key per
// level, in NamedLevels or FreshLevels.
//
// A name's addressed keys are the keys some read of it lands on or passes, from the level its
// binding lands on onward; a load gathers them in AddressedKeys. When a template a pick renders
// reads a field, the pick keeps that read's value only if the read's first-level key, put under
// the rendering template's key, is one of the name's addressed keys. LandsInPick decides this.
//
// In geo.SE.address, {.locality as l} binds l, and {l.street.name} passes the levels "", "street"
// and "street.name", the last of which is its key. With {l.name} and {l.postal-code.code}, l's
// addressed keys are those three, "name", "postal-code" and "postal-code.code".
type PickKey string

// Under puts rel beneath k, so PickKey("a").Under("b") is "a.b". Here k is the key of what renders
// and rel a key from a fresh read's levels.
// It never returns k itself: on a render k is the key the render's env holds, and a memo keeping it
// would move every render's env to the heap.
func (k PickKey) Under(rel PickKey) PickKey {
	if k == "" {
		return rel
	}
	return k + "." + rel
}

// NamedLevels is the levels of a read through a name, whose path runs from its target's start: the
// key of that start, "", then the key of each longer prefix of path, shortest first.
func NamedLevels(path []string) []PickKey {
	levels := make([]PickKey, len(path)+1)
	for i := range levels {
		levels[i] = PickKey(grammar.JoinSegments(path[:i]))
	}
	return levels
}

// FreshLevels is the levels of a fresh read of head and segs: the key of head, then the key of each
// longer path, shortest first. It has no "" level, since the template the read sits in is already
// rendering and head is the first node the read can draw.
func FreshLevels(head string, segs []string) []PickKey {
	return NamedLevels(append([]string{head}, segs...))[1:]
}

// AddressedKeys is every key the reads of each name land on or pass, from the name's own level,
// with the spelling of the first read reaching it.
func AddressedKeys(reads map[*NameBinding][]*Read, l *LoadFinds) map[*NameBinding]map[PickKey]string {
	keys := map[*NameBinding]map[PickKey]string{}
	for b, rs := range reads {
		keys[b] = map[PickKey]string{}
		for _, r := range rs {
			for _, key := range r.Levels()[len(l.Binding(b).Target().Tail()):] {
				if _, seen := keys[b][key]; !seen {
					keys[b][key] = r.Spelling()
				}
			}
		}
	}
	return keys
}

// LevelsUnder is the levels of a as a pick keys them: each of a's levels, moved under at. a is a
// fresh read in a template the pick renders, and at is that template's key. It is the one rule that lines a fresh read's
// keys up with a named read's, and load and render both call it.
//
// The two kinds of read share a key because of where their levels start. A named read starts at
// the node its name's target starts at, and may draw a variant there, so its levels start at that
// node, keyed "" (NamedLevels). A fresh read starts in the template it sits in, which is already
// rendering. The first node it can draw is its head, so its levels start at the head (FreshLevels).
// at is that template's path from the target's start, so moved under at, a fresh read's key spells
// the path a named read's does. Where p binds a category whose format reads {first}, the {first}
// that {p} renders at "" and the {p.first} beside it both key "first", and read one draw.
func LevelsUnder(at PickKey, a Read) []PickKey {
	levels := make([]PickKey, len(a.levels))
	for i, l := range a.levels {
		levels[i] = at.Under(l)
	}
	return levels
}

// LandsInPick reports whether a pick keeps the value of a. a is a read in a template the pick
// renders, at is that template's key, and by is the spelling of the read of the name that addresses
// a. Load and render both call it, so they agree on what a pick keeps. A reference draws afresh and
// a read through a name reads its own name's pick, so the pick keeps neither. It keeps any other
// read where a read of the name addresses the first of LevelsUnder's keys, built here alone to spare the slice.
func LandsInPick(addressed map[PickKey]string, at PickKey, a Read) (by string, lands bool) {
	if grammar.IsRef(a.Head()) || a.Kind() == NamedRead {
		return "", false
	}
	by, lands = addressed[at.Under(a.levels[0])]
	return by, lands
}
