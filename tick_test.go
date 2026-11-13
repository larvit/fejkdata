package fejkdata

import (
	"slices"
	"strings"
	"testing"
	"time"
)

const tickLayout = "2006-01-02T15:04:05Z"

func drawTicks(t *testing.T, f *Generator, tmpl string, n int) []time.Time {
	t.Helper()
	ticks := make([]time.Time, n)
	for i := range ticks {
		at, err := time.Parse(tickLayout, mustRender(t, f, tmpl))
		if err != nil {
			t.Fatal(err)
		}
		ticks[i] = at
	}
	return ticks
}

func TestTickPrintsRowNAtTheStartPlusNStepsAndWithinTheJitter(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, at := range drawTicks(t, readOn(t, testDay), `"{tick(2026-01-01T00:00:00Z,30,5,'2006-01-02T15:04:05Z')}"`, 200) {
		if low := start.Add(time.Duration(i) * 30 * time.Second); at.Before(low) || at.After(low.Add(5*time.Second)) {
			t.Fatalf("row %d = %s, want %s to %s", i+1, at.Format(tickLayout), low.Format(tickLayout), low.Add(5*time.Second).Format(tickLayout))
		}
	}
}

func TestTickRowsAscendWhenTheJitterIsOneBelowTheStep(t *testing.T) {
	ticks := drawTicks(t, readOn(t, testDay), `"{tick(2026-01-01T00:00:00Z,30,29,'2006-01-02T15:04:05Z')}"`, 5000)
	for i := 1; i < len(ticks); i++ {
		if !ticks[i].After(ticks[i-1]) {
			t.Fatalf("row %d = %s does not come after row %d = %s", i+1, ticks[i], i, ticks[i-1])
		}
	}
}

func TestTickDrawsTheJittersFirstAndLastSecond(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	seen := map[time.Duration]bool{}
	for i, at := range drawTicks(t, readOn(t, testDay), `"{tick(2026-01-01T00:00:00Z,30,5,'2006-01-02T15:04:05Z')}"`, 600) {
		seen[at.Sub(start)-time.Duration(i)*30*time.Second] = true
	}
	if !seen[0] || !seen[5*time.Second] || len(seen) != 6 {
		t.Errorf("jitter 5 drew the offsets %v, want each of 0 to 5 seconds and no other", seen)
	}
}

func TestTickRowsKeepAscendingPastTwoHundredNinetyTwoYears(t *testing.T) {
	ticks := drawTicks(t, readOn(t, testDay), `"{tick(2026-01-01,86400,0,'2006-01-02T15:04:05Z')}"`, 107000)
	for i := 1; i < len(ticks); i++ {
		if !ticks[i].After(ticks[i-1]) {
			t.Fatalf("row %d = %s does not come after row %d = %s", i+1, ticks[i], i, ticks[i-1])
		}
	}
}

func TestTickWithNoJitterPrintsExactSteps(t *testing.T) {
	f := readOn(t, testDay)
	for _, want := range []string{"2026-01-01 00:00:00", "2026-01-01 00:01:00", "2026-01-01 00:02:00"} {
		if got := mustRender(t, f, `"{tick(2026-01-01,60,0,'2006-01-02 15:04:05')}"`); got != want {
			t.Fatalf("tick = %q, want %q: a day is read as its midnight, UTC", got, want)
		}
	}
}

func TestTickRepeatsForTheSameSeed(t *testing.T) {
	const tmpl = `"{tick(2026-01-01T00:00:00Z,30,5,'2006-01-02T15:04:05Z')}"`
	a := drawTicks(t, readOn(t, testDay), tmpl, 50)
	b := drawTicks(t, readOn(t, testDay), tmpl, 50)
	if !slices.Equal(a, b) {
		t.Fatalf("two generators on one seed drew %v and %v", a, b)
	}
}

func TestTickCallsWithTheSameArgumentsStepOneCounter(t *testing.T) {
	got := mustRender(t, readOn(t, testDay), `"{tick(2026-01-01,60,0,'15:04')} {tick(2026-01-01,60,0,'15:04')}"`)
	if got != "00:00 00:01" {
		t.Fatalf("two equal calls = %q, want them to step one counter", got)
	}
}

func TestTickCallsWithOtherArgumentsStepApart(t *testing.T) {
	got := mustRender(t, readOn(t, testDay), `"{tick(2026-01-01,60,0,'15:04')} {tick(2026-01-01,30,0,'15:04')}"`)
	if got != "00:00 00:00" {
		t.Fatalf("two calls of other steps = %q, want each to start at its first row", got)
	}
}

func TestTickRefusesWhatCouldSwapRowsOrIsNoTime(t *testing.T) {
	for name, c := range map[string]struct{ tmpl, want string }{
		"jitter equal to step": {`"{tick(2026-01-01,30,30,'15:04')}"`, "jitter 30 is not below step 30"},
		"jitter above step":    {`"{tick(2026-01-01,30,31,'15:04')}"`, "jitter 31 is not below step 30"},
		"no step":              {`"{tick(2026-01-01,0,0,'15:04')}"`, "step"},
		"step above a day":     {`"{tick(2026-01-01,86401,0,'15:04')}"`, "step"},
		"negative jitter":      {`"{tick(2026-01-01,30,-1,'15:04')}"`, "jitter"},
		"fractional step":      {`"{tick(2026-01-01,1.5,0,'15:04')}"`, "step"},
		"start not a time":     {`"{tick(soon,30,5,'15:04')}"`, "start"},
		"start not a zone":     {`"{tick(2026-01-01T00:00:00,30,5,'15:04')}"`, "start"},
		"unquoted layout":      {`"{tick(2026-01-01,30,5,15:04)}"`, "'15:04'"},
		"too few arguments":    {`"{tick(2026-01-01,30,5)}"`, "4 arguments"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resolved(t, c.tmpl); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("compile(%s) = %v, want an error mentioning %q", c.tmpl, err, c.want)
			}
		})
	}
}

func TestDateDrawsBetweenDaysRelativeToTheCallersToday(t *testing.T) {
	f := readOn(t, dayOf(2026, 11, 12))
	first, last := "9999-12-31", "0000-01-01"
	for range 3000 {
		got := mustRender(t, f, `"{date(today-30,today+14,'2006-01-02')}"`)
		first, last = min(first, got), max(last, got)
	}
	if first != "2026-10-13" || last != "2026-11-26" {
		t.Errorf("date(today-30,today+14) on 2026-11-12 drew %s to %s, want 2026-10-13 to 2026-11-26", first, last)
	}
}

func TestDateOfTodayAloneIsTheCallersDay(t *testing.T) {
	for _, want := range []string{"2026-02-28", "2027-03-01"} {
		t.Run(want, func(t *testing.T) {
			today, err := time.Parse("2006-01-02", want)
			if err != nil {
				t.Fatal(err)
			}
			if got := mustRender(t, readOn(t, today), `"{date(today,today,'2006-01-02')}"`); got != want {
				t.Errorf("date(today,today) on %s = %q, want %q", want, got, want)
			}
		})
	}
}

func TestDateKeepsARelativeBoundInsideYears0000To9999(t *testing.T) {
	for name, c := range map[string]struct {
		today      time.Time
		tmpl, want string
	}{
		"far past":   {dayOf(2026, 11, 12), `"{date(today-1000000,today-1000000,'2006-01-02')}"`, "0000-01-01"},
		"far future": {dayOf(9999, 12, 31), `"{date(today+1,today+1,'2006-01-02')}"`, "9999-12-31"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := mustRender(t, readOn(t, c.today), c.tmpl); got != c.want {
				t.Errorf("%s on %s = %q, want %q", c.tmpl, c.today.Format("2006-01-02"), got, c.want)
			}
		})
	}
}

func TestDateAcceptsARelativeOffsetOfAMillionDays(t *testing.T) {
	if _, err := resolved(t, `"{date(today-1000000,today+1000000,'2006-01-02')}"`); err != nil {
		t.Errorf("a bound of a million days either way is refused: %v", err)
	}
}

func TestAddDaysReadsADateRelativeToToday(t *testing.T) {
	got := mustRender(t, readOn(t, dayOf(2026, 11, 12)), `{"format":"{x as s}{adddays(s,1)}","x":"{date(today-1,today-1,'2006-01-02')}"}`)
	if got != "2026-11-12" {
		t.Errorf("adddays over date(today-1,today-1) on 2026-11-12 = %q, want 2026-11-12", got)
	}
}

func TestDateRefusesRelativeBoundsThatCannotBeOrdered(t *testing.T) {
	for name, c := range map[string]struct{ tmpl, want string }{
		"a day and a relative bound": {`"{date(today-30,2026-12-31,'2006-01-02')}"`, "both days or both relative to today"},
		"a relative bound and a day": {`"{date(2026-01-01,today+1,'2006-01-02')}"`, "both days or both relative to today"},
		"from after to":              {`"{date(today+14,today-30,'2006-01-02')}"`, "is after"},
		"an offset that is no count": {`"{date(today-x,today,'2006-01-02')}"`, "today-x"},
		"an offset past a million":   {`"{date(today-1000001,today,'2006-01-02')}"`, "exceeds"},
		"a future offset past that":  {`"{date(today,today+1000001,'2006-01-02')}"`, "exceeds"},
		"an offset with no sign":     {`"{date(today5,today,'2006-01-02')}"`, "today5"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resolved(t, c.tmpl); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("compile(%s) = %v, want an error mentioning %q", c.tmpl, err, c.want)
			}
		})
	}
}
