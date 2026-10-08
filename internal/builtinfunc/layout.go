package builtinfunc

import (
	"fmt"
	"strings"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
)

const dayLayout = "2006-01-02"

// quotedLayout reports whether an arg carries the single quotes a layout is written in.
func quotedLayout(a string) bool {
	return len(a) >= 2 && a[0] == '\'' && a[len(a)-1] == '\''
}

// layoutArg is the Go layout a quoted arg holds, refused when unquoted.
// docs/decisions.md#a-layout-is-always-quoted
func layoutArg(a string) (string, error) {
	if !quotedLayout(a) {
		bare := strings.Trim(a, `'"`)
		if strings.HasPrefix(a, `"`) || strings.HasSuffix(a, `"`) {
			return "", fmt.Errorf("layout %s is double-quoted; write '%s'", a, bare)
		}
		return "", fmt.Errorf("layout %s is not quoted; write '%s'", a, bare)
	}
	return a[1 : len(a)-1], nil
}

// layoutOf is layoutArg for an arg a check already validated.
func layoutOf(a string) string {
	layout, err := layoutArg(a)
	if err != nil {
		panic(invariant.Broken("builtin arg %q reached prep unvalidated: %v", a, err))
	}
	return layout
}

// layoutArity checks a call ending in a layout takes n args, naming the quoted
// layout where an unquoted one split into more.
func layoutArity(name string, n int, a []string) error {
	if len(a) == n {
		return nil
	}
	hint := ""
	if len(a) > n && !holdsQuotedLayout(a[n-1:]) {
		hint = fmt.Sprintf("; a layout holding a comma is quoted: '%s'", strings.Trim(strings.Join(a[n-1:], ", "), `'"`))
	}
	return fmt.Errorf("%s takes %d argument%s, got %d%s", name, n, plural(n), len(a), hint)
}

// holdsQuotedLayout reports whether the surplus args already carry a quoted layout,
// which no comma split apart.
func holdsQuotedLayout(a []string) bool {
	for _, arg := range a {
		if quotedLayout(arg) {
			return true
		}
	}
	return false
}
func dateArgs(a []string) error {
	if err := layoutArity("date", 3, a); err != nil {
		return err
	}
	from, err := time.Parse(dayLayout, a[0])
	if err != nil {
		return fmt.Errorf("date(from,to,layout): from %q is not a YYYY-MM-DD date", a[0])
	}
	to, err := time.Parse(dayLayout, a[1])
	if err != nil {
		return fmt.Errorf("date(from,to,layout): to %q is not a YYYY-MM-DD date", a[1])
	}
	if to.Before(from) {
		return fmt.Errorf("date(from,to,layout): from %s is after to %s", a[0], a[1])
	}
	_, err = layoutArg(a[2])
	return err
}
func timeArg(a []string) error {
	if err := layoutArity("time", 1, a); err != nil {
		return err
	}
	layout, err := layoutArg(a[0])
	if err == nil && namesADateField(layout) {
		return fmt.Errorf("time(layout): '%s' names a date field, and time draws no date; write date(from,to,layout)", layout)
	}
	return err
}

// namesADateField reports a layout that prints two instants differing in their date alone apart.
func namesADateField(layout string) bool {
	day, nextYear := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC), time.Date(2010, 11, 12, 4, 5, 6, 0, time.UTC)
	return day.Format(layout) != nextYear.Format(layout)
}

// datePrep draws a second in [from 00:00:00, to 23:59:59] UTC; the span is counted
// in seconds, since a Duration overflows past 292 years.
func datePrep(a []string) Call {
	from, err := time.Parse(dayLayout, a[0])
	if err != nil {
		panic(invariant.Broken("builtin arg %q reached prep unvalidated: %v", a[0], err))
	}
	to, err := time.Parse(dayLayout, a[1])
	if err != nil {
		panic(invariant.Broken("builtin arg %q reached prep unvalidated: %v", a[1], err))
	}
	layout, start, span := layoutOf(a[2]), from.Unix(), int(to.Unix()-from.Unix())+86400
	return func(s *drawstate.State, _ string, _ []string) string {
		return time.Unix(start+int64(s.IntN(span)), 0).UTC().Format(layout)
	}
}
func timePrep(a []string) Call {
	layout := layoutOf(a[0])
	return func(s *drawstate.State, _ string, _ []string) string {
		return time.Unix(int64(s.IntN(86400)), 0).UTC().Format(layout)
	}
}
