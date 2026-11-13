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
	if !slices.IsSortedFunc(ticks, func(a, b time.Time) int { return a.Compare(b) }) {
		t.Fatal("rows swapped order")
	}
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
	for today, want := range map[time.Time]string{dayOf(2026, 2, 28): "2026-02-28", dayOf(2027, 3, 1): "2027-03-01"} {
		if got := mustRender(t, readOn(t, today), `"{date(today,today,'2006-01-02')}"`); got != want {
			t.Errorf("date(today,today) on %s = %q, want %q", want, got, want)
		}
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
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resolved(t, c.tmpl); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("compile(%s) = %v, want an error mentioning %q", c.tmpl, err, c.want)
			}
		})
	}
}
