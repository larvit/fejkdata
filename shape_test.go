package fejkdata

import (
	"testing"

	"github.com/larvit/fejkdata/internal/shipped"
)

const repin = `repin: docker compose run --rm --user "$(id -u):$(id -g)" repin`

func TestShippedShapeIsPinned(t *testing.T) {
	generatedFileIsCurrent(t, shipped.Shape(&newShippedWhole(t).root), "a removed, renamed or retyped line is a breaking change; add the CHANGELOG.md entry, then "+repin)
}

func TestTheListingsShapeIsPinned(t *testing.T) {
	generatedFileIsCurrent(t, shipped.ListingShape(), "a removed or renamed line is a major, as the README's Versioning says; add the CHANGELOG.md entry, then "+repin)
}

func TestShippedShapeNamesReads(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{
		"a":     `{"format":"{x}","x":["{/b}",{"format":"{/c.v}","weight":2}]}`,
		"b":     `"y"`,
		"c":     `{"format":"{v} {n}","n":[null,{"format":"{int(1,9)}","datatype":"integer"}],"v":["z","w"]}`,
		"d/pos": `["{.q}","{/b}"]`,
		"d/q":   `"r"`,
		"e":     `"{/a}"`,
	}))
	want := "a\tformat \"{x}\"\treads b c\na.x\tstring\nb\tformat \"y\"\nc\tformat \"{v} {n}\"\nc.n\tinteger null\nc.v\tstring\nd.pos\treads b d.q\nd.q\tformat \"r\"\ne\tformat \"{/a}\"\treads a\n"
	if got := string(shipped.Shape(&f.root).Body); got != want {
		t.Fatalf("Shape =\n%s\nwant\n%s", got, want)
	}
}

func TestShippedShapeNamesTables(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"w.json": `"x"`,
		"r.json": `{"format":"{name}","rows":"r.tsv","key":"code","name":"name","weight":"w"}`,
		"r.tsv":  "code\tname\tw\n1\ta\t2\n2\tb\t3\n",
		"m.json": `{"format":"{name} {/w}","rows":"m.tsv","key":"code","parent":"r"}`,
		"m.tsv":  "code\tname\tr\n10\tc\t1\n20\td\t2\n",
	}))
	want := "m\tformat \"{name} {/w}\"\tkey code\tparent r\treads w\nm.code\tstring\nm.name\tstring\nm.r\tstring\nr\tformat \"{name}\"\tkey code\tname name\tweight w\nr.code\tstring\nr.m\nr.m.code\nr.m.name\nr.m.r\nr.name\tstring\nr.w\tstring\nw\tformat \"x\"\n"
	if got := string(shipped.Shape(&f.root).Body); got != want {
		t.Fatalf("Shape =\n%s\nwant\n%s", got, want)
	}
}
