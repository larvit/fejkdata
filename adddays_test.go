package fejkdata

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAddDaysFallsNDaysAfterTheDateItReads(t *testing.T) {
	f := engine(1)
	tmpl := `{"format":"{start as s}{len as n}{s},{n},{adddays(s,n)}","start":"{date(2026-01-01,2026-12-31,'2006-01-02')}","len":"{int(1,30)}"}`
	eachDraw(200, func() string { return mustRender(t, f, tmpl) }, func(render string) {
		got := strings.Split(render, ",")
		start, err := time.Parse("2006-01-02", got[0])
		if err != nil || len(got) != 3 {
			t.Fatalf("render = %q, want start,n,until", got)
		}
		n, _ := strconv.Atoi(got[1])
		if want := start.AddDate(0, 0, n).Format("2006-01-02"); got[2] != want {
			t.Fatalf("render = %q, want until %s, %d days after %s", got, want, n, got[0])
		}
	})
}

const addDaysFromATimestamp = `{"format":"{created as c}{c},{adddays(c,gap)},{adddays(c,-3,'January 2, 2006')}","created":"{date(2026-12-01,2026-12-31,'2006-01-02T15:04:05Z')}","gap":"{int(1,30)}"}`

func TestAddDaysKeepsTheTimeOfDay(t *testing.T) {
	f := engine(2)
	eachDraw(200, func() string { return mustRender(t, f, addDaysFromATimestamp) }, func(render string) {
		got := strings.Split(render, ",")
		created, err := time.Parse(time.RFC3339, got[0])
		if err != nil || len(got) != 4 {
			t.Fatalf("render = %q, want created,updated,earlier", got)
		}
		updated, err := time.Parse(time.RFC3339, got[1])
		if gap := updated.Sub(created); err != nil || gap%(24*time.Hour) != 0 || gap < 24*time.Hour || gap > 30*24*time.Hour {
			t.Fatalf("render = %q, want updated 1 to 30 whole days after created", got)
		}
	})
}

func TestAddDaysPrintsInTheLayoutGiven(t *testing.T) {
	f := engine(2)
	eachDraw(200, func() string { return mustRender(t, f, addDaysFromATimestamp) }, func(render string) {
		got := strings.Split(render, ",")
		created, err := time.Parse(time.RFC3339, got[0])
		if err != nil || len(got) != 4 {
			t.Fatalf("render = %q, want created,updated,earlier", got)
		}
		if want := created.AddDate(0, 0, -3).Format("January 2") + ", " + created.AddDate(0, 0, -3).Format("2006"); got[2]+","+got[3] != want {
			t.Fatalf("render = %q, want %q three days before created", got, want)
		}
	})
}

func TestAddDaysReadsABornDate(t *testing.T) {
	f := engineOn(3, time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC))
	got := mustRender(t, f, `{"format":"{birth as b}{b} {adddays(b,1)}","birth":"{born(30,30,'20060102')}"}`)
	birth, err := time.Parse("20060102", got[:8])
	if err != nil || got[9:] != birth.AddDate(0, 0, 1).Format("20060102") {
		t.Errorf("render = %q, want a birth date and the day after it", got)
	}
}

func TestAddDaysRefusesWhatIsNoProvenDateOrDayCount(t *testing.T) {
	for tmpl, want := range map[string]string{
		`{"format":"{x as s}{adddays(s,1)}","x":"hello"}`:                                                                   "date(from,to,'layout') or born(min,max,'layout')",
		`{"format":"{x as s}{adddays(s,1)}","x":"on {date(2020-01-01,2020-12-31,'2006-01-02')}"}`:                           "date(from,to,'layout') or born(min,max,'layout')",
		`{"format":"{x as s}{adddays(s,1)}","x":"{date(2020-01-01,2020-12-31,'01-02')}"}`:                                   "does not read back",
		`{"format":"{x as s}{adddays(s,1)}","x":"{date(1950-01-01,2020-12-31,'060102')}"}`:                                  "does not read back",
		`{"format":"{x as s}{adddays(s,1)}","x":["{date(2020-01-01,2020-12-31,'2006-01-02')}","{born(1,2,'20060102')}"]}`:   "two layouts",
		`{"format":"{x as s}{adddays(s,n)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}","n":"{float(1,2,1)}"}`:         "whole number of days",
		`{"format":"{x as s}{adddays(s,n)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}","n":"many"}`:                   "whole number of days",
		`{"format":"{x as s}{adddays(s,2000000)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`:                        "exceeds",
		`{"format":"{x as s}{adddays(s,n)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}","n":"{int(1,2000000)}"}`:       "exceeds",
		`{"format":"{x as s}{y as u}{adddays(u,1)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}","y":"{adddays(s,1)}"}`: "one adddays",
		`{"format":"{x as s}{adddays(s,1,2006-01-02)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`:                   "'2006-01-02'",
		`{"format":"{x as s}{adddays(s)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`:                                "2 or 3 arguments",
		`{"format":"{x as s}{adddays(s,1)}","x":"{date(2020-01-01,2020-12-31,'2006-01')}"}`:                                 "does not read back",
		`{"format":"{x as s}{adddays(s,1)}","x":"{date(2020-01-01,2020-12-31,'200612')}"}`:                                  "does not read back",
		`{"format":"{x as s}{adddays(s,1)}","x":"{date(2020-01-01,2020-12-31,'2006-1_2')}"}`:                                "does not read back",
		`{"format":"{x as s}{adddays(s,1)}","x":"{date(2020-01-01,2020-12-31,'Jan3_2 2006')}"}`:                             "does not read back",
		`{"format":"{x as s}{adddays(s,-)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`:                              "not an integer",
		`{"format":"{x as s}{adddays(s,1000001)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`:                        "exceeds",
		`{"format":"{x as s}{adddays(s,-1000001)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`:                       "exceeds",
		`{"format":"{x as s}{adddays(s,n)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}","n":"{int(1,1000001)}"}`:       "1000001",
		`{"format":"{x as s}{adddays(s,-n)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}","n":"{int(1,9)}"}`:            "draw it negative",
	} {
		t.Run(tmpl, func(t *testing.T) {
			_, err := resolved(t, tmpl)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("compile(%s) = %v, want an error mentioning %q", tmpl, err, want)
			}
		})
	}
}

func TestAddDaysAcceptsAMillionDaysEitherWay(t *testing.T) {
	for _, tmpl := range []string{
		`{"format":"{x as s}{adddays(s,1000000)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`,
		`{"format":"{x as s}{adddays(s,-1000000)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}"}`,
		`{"format":"{x as s}{adddays(s,n)}","x":"{date(2020-01-01,2020-12-31,'2006-01-02')}","n":"{int(-1000000,1000000)}"}`,
	} {
		t.Run(tmpl, func(t *testing.T) {
			if _, err := resolved(t, tmpl); err != nil {
				t.Errorf("compile(%s) = %v, want no error", tmpl, err)
			}
		})
	}
}
