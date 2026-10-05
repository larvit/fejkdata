package rows

import (
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/drawstate"
)

func parse(t *testing.T, segment, data string, o Options) *Table[string] {
	t.Helper()
	tb, err := Parse(segment, segment, "g."+segment, segment+".tsv", data, o)
	if err != nil {
		t.Fatalf("Parse(%s): %v", segment, err)
	}
	return tb
}

// family is a country table, SE and NO, and a city table linked to it, where Borg names a city in each country.
func family(t *testing.T) (country, city *Table[string]) {
	t.Helper()
	country = parse(t, "country", "code\tname\nSE\tSweden\nNO\tNorway\n", Options{Key: "code"})
	city = parse(t, "city", "country\tname\nSE\tBorg\nNO\tBorg\nSE\tLund\n", Options{Name: "name", Parent: "country"})
	if err := city.Link(country, func(string) *Table[string] { return nil }); err != nil {
		t.Fatalf("Link: %v", err)
	}
	return country, city
}

func TestParseRefuses(t *testing.T) {
	for _, c := range []struct {
		name, data string
		o          Options
		want       string
	}{
		{"no header", "", Options{}, "no header line"},
		{"one row", "a\n1\n", Options{}, "has one row"},
		{"a short row", "a\tb\n1\t2\n3\n", Options{}, "line 3 has fewer cells"},
		{"a repeated key", "k\nx\nx\n", Options{Key: "k"}, `key "x" repeats line 2`},
		{"the key named as name", "k\nx\ny\n", Options{Key: "k", Name: "k"}, "name names the key column"},
		{"a weight of zero", "k\tw\nx\t1\ny\t0\n", Options{Weight: "w"}, `weight "0" is not a positive number`},
		{"a key a selector cannot spell", "k\nx\ny|z\n", Options{Key: "k"}, `contains "|"`},
		{"an option naming no column", "k\nx\ny\n", Options{Key: "nope"}, `key names no column "nope"`},
	} {
		if _, err := Parse("o", "t", "g.t", "t.tsv", c.data, c.o); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Parse = %v, want %q", c.name, err, c.want)
		}
	}
}

func TestLinkRefusesAParentRowNoChildLinksTo(t *testing.T) {
	country := parse(t, "country", "code\nSE\nNO\n", Options{Key: "code"})
	city := parse(t, "city", "country\tname\nSE\tLund\nSE\tBorg\n", Options{Parent: "country"})
	if err := city.Link(country, func(string) *Table[string] { return nil }); err == nil || !strings.Contains(err.Error(), `country "NO"`) {
		t.Errorf("Link = %v, want the unlinked row NO named", err)
	}
	if city.Parent() != nil {
		t.Error("a refused link set the parent")
	}
}

func TestOwnerIsWhatParseWasGiven(t *testing.T) {
	if _, city := family(t); city.Parent().Owner() != "country" {
		t.Errorf("Owner = %q, want country", city.Parent().Owner())
	}
}

func TestSelectResolvesANameInsideThePinnedParent(t *testing.T) {
	country, city := family(t)
	var p Pins[string]
	if _, err := p.Select(city, "Borg"); err == nil || !strings.Contains(err.Error(), "names 2 rows") {
		t.Fatalf("Select(Borg) with no country pinned = %v, want it ambiguous", err)
	}
	if _, err := p.Select(country, "NO"); err != nil {
		t.Fatal(err)
	}
	r, err := p.Select(city, "Borg")
	if err != nil || r != 1 {
		t.Fatalf("Select(Borg) inside NO = %d, %v, want row 1", r, err)
	}
	if err := p.PinRow(city, 2); err == nil || !strings.Contains(err.Error(), "two rows of g.city") {
		t.Errorf("PinRow(Lund) over Borg = %v, want a clash", err)
	}
}

func TestPinRowRefusesARowOutsideThePinnedParent(t *testing.T) {
	country, city := family(t)
	var p Pins[string]
	if err := p.PinRow(country, 1); err != nil {
		t.Fatal(err)
	}
	if err := p.PinRow(city, 2); err == nil || !strings.Contains(err.Error(), "is not inside g.country[NO]") {
		t.Errorf("PinRow(Lund) inside NO = %v, want it refused", err)
	}
}

func TestDrawInDrawsInsideThePinnedParentAndPinsItsAncestors(t *testing.T) {
	country, city := family(t)
	s := drawstate.New(1)
	for i := 0; i < 50; i++ {
		var p Pins[string]
		if err := p.PinRow(country, 0); err != nil {
			t.Fatal(err)
		}
		if r := city.DrawIn(s, &p); city.Cell(r, 0) != "SE" {
			t.Fatalf("DrawIn inside SE drew a city of %s", city.Cell(r, 0))
		}
		var q Pins[string]
		r := city.DrawIn(s, &q)
		if pr := q.MustRow(country); country.Cell(pr, 0) != city.Cell(r, 0) {
			t.Fatalf("DrawIn pinned country %s above a city of %s", country.Cell(pr, 0), city.Cell(r, 0))
		}
		if _, ok := q.Above(country).Pinned(city); ok {
			t.Fatal("Above(country) kept the city's pin")
		}
	}
}

func TestProveStepUpNamesTheParent(t *testing.T) {
	country, city := family(t)
	if err := city.ProveStepUp([]string{"country", "name"}); err != nil {
		t.Errorf("ProveStepUp(country.name) = %v", err)
	}
	if err := city.ProveStepUp([]string{"name"}); err == nil || !strings.Contains(err.Error(), "..country") {
		t.Errorf("ProveStepUp(name) = %v, want ..country named", err)
	}
	if err := country.ProveStepUp([]string{"x"}); err == nil {
		t.Error("ProveStepUp from a table with no parent passed")
	}
}
