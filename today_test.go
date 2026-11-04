package fejkdata

import (
	"strings"
	"testing"
	"time"
)

func dayOf(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func readOn(t *testing.T, today time.Time) *Generator {
	t.Helper()
	f, err := New(WithDataFS(), WithSeed(1), WithToday(today))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestNewWithoutADateFailsNamingWithToday(t *testing.T) {
	for name, opts := range map[string][]Option{
		"no date":   {WithDataFS(), WithSeed(1)},
		"zero date": {WithDataFS(), WithSeed(1), WithToday(time.Time{})},
	} {
		if _, err := New(opts...); err == nil || !strings.Contains(err.Error(), "WithToday(") {
			t.Errorf("New(%s) = %v, want it refused, naming WithToday", name, err)
		}
	}
}

func TestNewRefusesADateOutsideYearsOneTo9999(t *testing.T) {
	for _, year := range []int{-1, 0, 10000} {
		_, err := New(WithDataFS(), WithSeed(1), WithToday(dayOf(year, 6, 15)))
		if err == nil || !strings.Contains(err.Error(), "WithToday") {
			t.Errorf("New(WithToday in year %d) = %v, want it refused, naming WithToday", year, err)
		}
	}
	for _, year := range []int{1, 9999} {
		if _, err := New(WithDataFS(), WithSeed(1), WithToday(dayOf(year, 6, 15))); err != nil {
			t.Errorf("New(WithToday in year %d) = %v, want it taken", year, err)
		}
	}
}

// A range of ages counts by calendar year, as the law's "the year a person turns 100" does.
func TestBornDrawsTheBirthYearsOfTheAges(t *testing.T) {
	f := readOn(t, dayOf(2026, 11, 4))
	first, last := "9999-12-31", "0000-01-01"
	for range 3000 {
		got := mustRender(t, f, `"{born(100,113,'2006-01-02')}"`)
		first, last = min(first, got), max(last, got)
	}
	if first < "1913-01-01" || first[:4] != "1913" || last > "1926-12-31" || last[:4] != "1926" {
		t.Errorf("born(100,113) read on 2026-11-04 drew %s to %s, want 1913-01-01 to 1926-12-31 at its ends", first, last)
	}
}

func TestBornDrawsNoDayAfterToday(t *testing.T) {
	f := readOn(t, dayOf(2026, 11, 4))
	seen := map[string]bool{}
	for range 5000 {
		got := mustRender(t, f, `"{born(0,0,'2006-01-02')}"`)
		if got < "2026-01-01" || got > "2026-11-04" {
			t.Fatalf("born(0,0) read on 2026-11-04 = %s, want a day from 2026-01-01 to 2026-11-04", got)
		}
		seen[got] = true
	}
	if !seen["2026-01-01"] || !seen["2026-11-04"] {
		t.Error("born(0,0) read on 2026-11-04 never drew the year's first day or the date itself")
	}
	jan1 := readOn(t, dayOf(2026, 1, 1))
	for range 50 {
		if got := mustRender(t, jan1, `"{born(0,0,'2006-01-02')}"`); got != "2026-01-01" {
			t.Fatalf("born(0,0) read on 2026-01-01 = %s, want that day", got)
		}
	}
}

// The date is the calendar day as written in its own location, whatever UTC says it is then.
func TestWithTodayReadsTheDayAsWritten(t *testing.T) {
	ahead := readOn(t, time.Date(2027, 1, 1, 0, 30, 0, 0, time.FixedZone("", 3600)))
	behind := readOn(t, time.Date(2026, 12, 31, 23, 30, 0, 0, time.FixedZone("", -5*3600)))
	for range 50 {
		if got := mustRender(t, ahead, `"{born(0,0,'2006-01-02')}"`); got != "2027-01-01" {
			t.Fatalf("born(0,0) read on 2027-01-01 at 00:30 +01:00 = %s, want 2027-01-01", got)
		}
		if got := mustRender(t, behind, `"{born(0,0,'2006')}"`); got != "2026" {
			t.Fatalf("born(0,0) read on 2026-12-31 at 23:30 -05:00 = %s, want a day in 2026", got)
		}
	}
	noon, midnight := readOn(t, time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)), readOn(t, dayOf(2026, 11, 4))
	for range 50 {
		a, b := mustRender(t, noon, `"{born(0,99,'2006-01-02T15:04:05')}"`), mustRender(t, midnight, `"{born(0,99,'2006-01-02T15:04:05')}"`)
		if a != b {
			t.Fatalf("read at noon %s, read at midnight %s, want the time of day to change nothing", a, b)
		}
	}
}

func TestTheDateRepeatsARunAndOnlyMovesWhatReadsIt(t *testing.T) {
	const template = `"{born(0,99,'2006-01-02')} {int(1,1000000)}"`
	a, again, later := readOn(t, dayOf(2026, 11, 4)), readOn(t, dayOf(2026, 11, 4)), readOn(t, dayOf(2040, 11, 4))
	moved := false
	for range 50 {
		x, y, z := mustRender(t, a, template), mustRender(t, again, template), mustRender(t, later, template)
		if x != y {
			t.Fatalf("one seed and date drew %q and %q", x, y)
		}
		if x[strings.Index(x, " "):] != z[strings.Index(z, " "):] {
			t.Fatalf("{int} read on two dates drew %q and %q, want the draw that reads no date unmoved", x, z)
		}
		moved = moved || x[:10] != z[:10]
	}
	if !moved {
		t.Error("born drew the same days read on two dates")
	}
}

func TestBornRefusesItsMistakes(t *testing.T) {
	f := readOn(t, dayOf(2026, 11, 4))
	for tmpl, want := range map[string]string{
		`"{born(5,1,'06')}"`:      "min 5",
		`"{born(-1,5,'06')}"`:     "negative",
		`"{born(a,5,'06')}"`:      "min",
		`"{born(0,b,'06')}"`:      "max",
		`"{born(0,10000,'06')}"`:  "9999",
		`"{born(0,5,06)}"`:        "not quoted",
		`"{born(0,5)}"`:           "3 arguments",
		`"{born(0,5,'06',0102)}"`: "3 arguments",
	} {
		_, err := f.FakeTemplate(tmpl)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s = %v, want it refused, naming %q", tmpl, err, want)
		}
	}
}

// luhn(n) reads only the last n digits printed, so a century can sit before them.
func TestLuhnOverTheLastDigits(t *testing.T) {
	f := engine(1)
	for tmpl, want := range map[string]string{
		`"19811218987{luhn(9)}"`:  "198112189876",
		`"19811218-987{luhn(9)}"`: "19811218-9876",
		`"811218987{luhn(9)}"`:    "8112189876",
		`"12{luhn(9)}"`:           "125",
		`"7992739871{luhn(10)}"`:  "79927398713",
		`"19811218987{luhn(11)}"`: "19811218987" + mustRender(t, f, `"19811218987{luhn()}"`)[11:],
	} {
		if got := mustRender(t, f, tmpl); got != want {
			t.Errorf("%s = %q, want %q", tmpl, got, want)
		}
	}
	for _, bad := range []string{`"{luhn(0)}"`, `"{luhn(-1)}"`, `"{luhn(x)}"`, `"{luhn(1,2)}"`, `"{luhn(1048577)}"`} {
		if _, err := f.FakeTemplate(bad); err == nil || !strings.Contains(err.Error(), "luhn") {
			t.Errorf("%s = %v, want it refused, naming luhn", bad, err)
		}
	}
}
