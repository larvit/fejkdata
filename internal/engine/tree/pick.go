package tree

// PickKey is the key a named pick's memo keeps one level's draw under.
//
// A level is one prefix of a read's path, and a read has one for each node it passes; a choice
// and the variant drawn from it, or a table and its row, share one. A level's key is its path from
// where the read starts, and landsInPick says where that is. Load gives each read one key per
// level, in namedLevels or freshLevels.
//
// A name's addressed keys are the keys some read of it lands on or passes, from the level its
// binding lands on onward; load gathers them in addressedKeys. When a template a pick renders
// reads a field, the pick keeps that read's value only if the read's first-level key, put under
// the rendering template's key, is one of the name's addressed keys. landsInPick decides this,
// and readUnder reads the kept value.
//
// In geo.SE.address, {.locality as l} binds l, and {l.street.name} passes the levels "", "street"
// and "street.name", the last of which is its key. With {l.name} and {l.postal-code.code}, l's
// addressed keys are those three, "name", "postal-code" and "postal-code.code".
type PickKey string

// Under puts rel beneath k, so PickKey("a").Under("b") is "a.b". Here k is the key of what renders
// and rel a key from a fresh read's levels.
// It never returns k itself: on a render k is env.pickAt, and a memo keeping it would move every
// render's env to the heap.
func (k PickKey) Under(rel PickKey) PickKey {
	if k == "" {
		return rel
	}
	return k + "." + rel
}
