package fejkdata

import (
	"fmt"
	"strings"
	"time"
)

const dayLayout = "2006-01-02"

// The instants a layout is proved against: layoutProbe2 is alike in no field, while
// layoutDay differs from layoutProbe in its date fields alone and layoutClock in its
// clock fields alone, so formatting two of them tells which kind a layout names.
var (
	layoutProbe  = time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	layoutProbe2 = time.Date(2010, 11, 12, 13, 14, 15, 0, time.UTC)
	layoutDay    = time.Date(2010, 11, 12, 4, 5, 6, 0, time.UTC)
	layoutClock  = time.Date(2001, 2, 3, 13, 14, 15, 0, time.UTC)
)

func namesAField(layout string) bool {
	return layoutProbe.Format(layout) != layoutProbe2.Format(layout)
}
func namesADateField(layout string) bool {
	return layoutProbe.Format(layout) != layoutDay.Format(layout)
}
func namesAClockField(layout string) bool {
	return layoutProbe.Format(layout) != layoutClock.Format(layout)
}

// quotedLayout reports whether an arg carries the single quotes a layout is written in.
func quotedLayout(a string) bool {
	return len(a) >= 2 && a[0] == '\'' && a[len(a)-1] == '\''
}

// layoutArg is the Go layout a quoted arg holds, refused when unquoted or constant.
// docs/decisions.md#a-layout-is-always-quoted
func layoutArg(a string) (string, error) {
	if !quotedLayout(a) {
		bare := strings.Trim(a, `'"`)
		if strings.HasPrefix(a, `"`) || strings.HasSuffix(a, `"`) {
			return "", fmt.Errorf("layout %s is double-quoted; write '%s'", a, bare)
		}
		return "", fmt.Errorf("layout %s is not quoted; write '%s'", a, bare)
	}
	layout := a[1 : len(a)-1]
	if !namesAField(layout) {
		return "", fmt.Errorf("layout '%s' names no field, so it is the constant %q; write it as text", layout, layout)
	}
	return layout, nil
}

// layoutOf is layoutArg for an arg a check already validated.
func layoutOf(a string) string {
	layout, err := layoutArg(a)
	if err != nil {
		panic(fmt.Sprintf("fejkdata: builtin arg %q reached prep unvalidated: %v", a, err))
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
func dateArgs(_ map[string]node, a []string) error {
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
	layout, err := layoutArg(a[2])
	if err != nil {
		return err
	}
	if !namesADateField(layout) {
		return fmt.Errorf("date(from,to,layout): '%s' names no date field; write time('%s')", layout, layout)
	}
	if from.Equal(to) && !namesAClockField(layout) {
		return fmt.Errorf("date(%s,%s,'%s') is the constant %q; write it as text", a[0], a[1], layout, from.Format(layout))
	}
	return nil
}
func timeArg(_ map[string]node, a []string) error {
	if err := layoutArity("time", 1, a); err != nil {
		return err
	}
	layout, err := layoutArg(a[0])
	if err != nil {
		return err
	}
	if namesADateField(layout) {
		return fmt.Errorf("time(layout): '%s' names a date field; write date(from,to,layout)", layout)
	}
	return nil
}

// datePrep draws a second in [from 00:00:00, to 23:59:59] UTC; the span is counted
// in seconds, since a Duration overflows past 292 years.
func datePrep(a []string) callFn {
	from, err := time.Parse(dayLayout, a[0])
	if err != nil {
		panic(fmt.Sprintf("fejkdata: builtin arg %q reached prep unvalidated: %v", a[0], err))
	}
	to, err := time.Parse(dayLayout, a[1])
	if err != nil {
		panic(fmt.Sprintf("fejkdata: builtin arg %q reached prep unvalidated: %v", a[1], err))
	}
	layout, start, span := layoutOf(a[2]), from.Unix(), int(to.Unix()-from.Unix())+86400
	return func(s *session, _ string, _ []string) string {
		return time.Unix(start+int64(s.IntN(span)), 0).UTC().Format(layout)
	}
}
func timePrep(a []string) callFn {
	layout := layoutOf(a[0])
	return func(s *session, _ string, _ []string) string {
		return time.Unix(int64(s.IntN(86400)), 0).UTC().Format(layout)
	}
}
