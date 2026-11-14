package rows

import (
	"fmt"
	"strings"
	"testing"
	"time"

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
		{"no rows", "a\n", Options{}, "has no rows below its header"},
		{"a short row", "a\tb\n1\t2\n3\n", Options{}, "line 3 has fewer cells"},
		{"a repeated key", "k\nx\nx\n", Options{Key: "k"}, `key "x" repeats line 2`},
		{"every weight zero", "k\tw\nx\t0\ny\t0\n", Options{Weight: "w"}, "every weight is 0"},
		{"a weight too close to 0", "k\tw\nx\t1\ny\t1e-400\n", Options{Weight: "w"}, "line 3: weight 1e-400 is too close to 0 to tell from it"},
		{"a negative weight", "k\tw\nx\t1\ny\t-1\n", Options{Weight: "w"}, `weight "-1" is not a number of 0 or more`},
		{"a key a selector cannot spell", "k\nx\ny|z\n", Options{Key: "k"}, `contains "|"`},
		{"an option naming no column", "k\nx\ny\n", Options{Key: "nope"}, `key names no column "nope"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Parse("o", "t", "g.t", "t.tsv", c.data, c.o); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("%s: Parse = %v, want %q", c.name, err, c.want)
			}
		})
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

func fiftyTimes(draw func()) {
	for i := 0; i < 50; i++ {
		draw()
	}
}

func TestDrawInDrawsInsideThePinnedParent(t *testing.T) {
	country, city := family(t)
	s := drawstate.New(1, time.Time{})
	fiftyTimes(func() {
		var p Pins[string]
		if err := p.PinRow(country, 0); err != nil {
			t.Fatal(err)
		}
		if r := city.DrawIn(s, &p); city.Cell(r, 0) != "SE" {
			t.Fatalf("DrawIn inside SE drew a city of %s", city.Cell(r, 0))
		}
	})
}

func TestDrawInPinsTheAncestorsOfTheRowItDraws(t *testing.T) {
	country, city := family(t)
	s := drawstate.New(1, time.Time{})
	fiftyTimes(func() {
		var q Pins[string]
		r := city.DrawIn(s, &q)
		if pr := q.MustRow(country); country.Cell(pr, 0) != city.Cell(r, 0) {
			t.Fatalf("DrawIn pinned country %s above a city of %s", country.Cell(pr, 0), city.Cell(r, 0))
		}
	})
}

func TestAboveDropsThePinsBelowATable(t *testing.T) {
	country, city := family(t)
	s := drawstate.New(1, time.Time{})
	fiftyTimes(func() {
		var q Pins[string]
		city.DrawIn(s, &q)
		if _, ok := q.Above(country).Pinned(city); ok {
			t.Fatal("Above(country) kept the city's pin")
		}
	})
}

func TestProveStepUpNamesTheParent(t *testing.T) {
	_, city := family(t)
	if err := city.ProveStepUp([]string{"country", "name"}); err != nil {
		t.Errorf("ProveStepUp(country.name) = %v", err)
	}
	if err := city.ProveStepUp([]string{"name"}); err == nil || !strings.Contains(err.Error(), "..country") {
		t.Errorf("ProveStepUp(name) = %v, want ..country named", err)
	}
}

func TestProveStepUpRefusesATableWithNoParent(t *testing.T) {
	country, _ := family(t)
	if err := country.ProveStepUp([]string{"x"}); err == nil {
		t.Error("ProveStepUp from a table with no parent passed")
	}
}

func TestParseRefusesAParentWhoseScatteredRowsAllWeighZero(t *testing.T) {
	data := "country\tcity\tw\nSE\tLund\t0\nNO\tBergen\t1\nSE\tBorg\t0\nDK\tAarhus\t0\nNO\tOslo\t2\n"
	_, err := Parse("o", "city", "g.city", "city.tsv", data, Options{Parent: "country", Weight: "w"})
	want := `city.tsv: every row under country "SE" weighs 0, so none is drawn there`
	if err == nil || err.Error() != want {
		t.Errorf("Parse = %v, want %q", err, want)
	}
}

func TestParseAcceptsAParentWhoseWeightIsSpreadOverScatteredRuns(t *testing.T) {
	data := "country\tcity\tw\nSE\tLund\t0\nNO\tBergen\t1\nSE\tBorg\t3\n"
	if _, err := Parse("o", "city", "g.city", "city.tsv", data, Options{Parent: "country", Weight: "w"}); err != nil {
		t.Errorf("Parse = %v, want SE's second run to carry its weight", err)
	}
}

func TestLinkNamesTheFirstLineOfAParentNoKeyHas(t *testing.T) {
	country := parse(t, "country", "code\nSE\nNO\n", Options{Key: "code"})
	city := parse(t, "city", "country\tname\nSE\tLund\nNO\tBergen\nXX\tA\nSE\tBorg\nYY\tB\n", Options{Parent: "country"})
	want := "city.tsv line 4: country \"XX\" is no key of country.tsv"
	if err := city.Link(country, func(string) *Table[string] { return nil }); err == nil || err.Error() != want {
		t.Errorf("Link = %v, want %q", err, want)
	}
}

func TestLinkNamesTheFirstParentRowNoChildLinksTo(t *testing.T) {
	country := parse(t, "country", "code\nSE\nNO\nDK\nFI\n", Options{Key: "code"})
	city := parse(t, "city", "country\tname\nSE\tLund\nNO\tBergen\nSE\tBorg\n", Options{Parent: "country"})
	want := "city.tsv links no row to country \"DK\"; every country row needs one, or drop line 4 of country.tsv"
	if err := city.Link(country, func(string) *Table[string] { return nil }); err == nil || err.Error() != want {
		t.Errorf("Link = %v, want %q", err, want)
	}
}

func TestDrawInReachesEveryScatteredRowOfTheParent(t *testing.T) {
	country := parse(t, "country", "code\nSE\nNO\n", Options{Key: "code"})
	city := parse(t, "city", "country\tname\tw\nSE\tLund\t1\nNO\tBergen\t1\nSE\tBorg\t1\nNO\tOslo\t0\nSE\tMalmö\t1\n", Options{Name: "name", Parent: "country", Weight: "w"})
	if err := city.Link(country, func(string) *Table[string] { return nil }); err != nil {
		t.Fatal(err)
	}
	s := drawstate.New(1, time.Time{})
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		var p Pins[string]
		if err := p.PinRow(country, 0); err != nil {
			t.Fatal(err)
		}
		r := city.DrawIn(s, &p)
		if city.Cell(r, 0) != "SE" {
			t.Fatalf("DrawIn inside SE drew a city of %s", city.Cell(r, 0))
		}
		seen[city.Cell(r, 1)] = true
	}
	if len(seen) != 3 {
		t.Errorf("DrawIn inside SE drew %v, want Lund, Borg and Malmö", seen)
	}
	for i := 0; i < 200; i++ {
		var p Pins[string]
		if err := p.PinRow(country, 1); err != nil {
			t.Fatal(err)
		}
		if r := city.DrawIn(s, &p); city.Cell(r, 1) != "Bergen" {
			t.Fatalf("DrawIn inside NO drew %s, whose weight is 0 or whose parent is SE", city.Cell(r, 1))
		}
	}
}

func parentTSV(parents, perParent int, scattered bool) (country, city string) {
	var c, t strings.Builder
	c.WriteString("code\n")
	t.WriteString("country\tname\tw\n")
	for p := 0; p < parents; p++ {
		fmt.Fprintf(&c, "P%d\n", p)
	}
	for i := 0; i < parents*perParent; i++ {
		p := i / perParent
		if scattered {
			p = i % parents
		}
		fmt.Fprintf(&t, "P%d\tn%d\t1\n", p, i)
	}
	return c.String(), t.String()
}

func BenchmarkParseAndLinkAParentColumn(b *testing.B) {
	for _, shape := range []struct {
		name      string
		scattered bool
	}{{"grouped", false}, {"scattered", true}} {
		b.Run(shape.name, func(b *testing.B) {
			countryData, cityData := parentTSV(2000, 50, shape.scattered)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				country, err := Parse("o", "country", "g.country", "country.tsv", countryData, Options{Key: "code"})
				if err != nil {
					b.Fatal(err)
				}
				city, err := Parse("o", "city", "g.city", "city.tsv", cityData, Options{Name: "name", Parent: "country", Weight: "w"})
				if err != nil {
					b.Fatal(err)
				}
				if err := city.Link(country, func(string) *Table[string] { return nil }); err != nil {
					b.Fatal(err)
				}
				var p Pins[string]
				if err := p.PinRow(country, 0); err != nil {
					b.Fatal(err)
				}
				city.DrawIn(drawstate.New(1, time.Time{}), &p)
			}
		})
	}
}
