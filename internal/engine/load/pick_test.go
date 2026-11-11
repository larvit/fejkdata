package load

import (
	"testing"

	"github.com/larvit/fejkdata/internal/engine/tree"
)

func TestAReadLandsInAPickWhereAReadOfItsNameAddressesItsFirstLevel(t *testing.T) {
	addressed := map[tree.PickKey]string{"first": "n.first", "k..x": "n.k..x", "w": "n.w", "x.first": "n.x.first"}
	for name, c := range map[string]struct {
		at     tree.PickKey
		a      tree.Read
		wantBy string
	}{
		"a field addressed":               {"", freshRead("first"), "n.first"},
		"a field addressed under at":      {"x", freshRead("first"), "n.x.first"},
		"a path whose first level is":     {"", freshRead("w.deeper"), "n.w"},
		"a field no read addresses":       {"", freshRead("last"), ""},
		"a field addressed elsewhere":     {"x", freshRead("w"), ""},
		"a reference a read's key spells": {"k", freshRead(".x"), ""},
		"a read through a name":           {"", tree.NewRead("", "n", nil, tree.NamedLevels([]string{"first"}), nil, nil).ThroughName(nil, 0), ""},
		"a read through a name at a key":  {"", tree.NewRead("", "first", nil, tree.FreshLevels("first", nil), nil, nil).ThroughName(nil, 0), ""},
	} {
		by, lands := tree.LandsInPick(addressed, c.at, c.a)
		if by != c.wantBy || lands != (c.wantBy != "") {
			t.Errorf("%s: LandsInPick = %q, %v, want %q", name, by, lands, c.wantBy)
		}
	}
}

func freshRead(spelling string) tree.Read {
	r := splitRead(spelling, nil)
	return tree.NewRead(spelling, r.head, r.tail, tree.FreshLevels(r.head, r.tail), nil, nil)
}
