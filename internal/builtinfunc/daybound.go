package builtinfunc

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// dayBound is a bound of a date call: a fixed day, or days from the caller's today.
type dayBound struct {
	day      time.Time
	days     int
	relative bool
}

// parseDayBound reads YYYY-MM-DD, or today, today+N, today-N.
func parseDayBound(arg string) (dayBound, error) {
	rest, relative := strings.CutPrefix(arg, "today")
	if !relative {
		day, err := time.Parse(dayLayout, arg)
		return dayBound{day: day}, err
	}
	if rest == "" {
		return dayBound{relative: true}, nil
	}
	days, err := strconv.Atoi(rest)
	if err != nil || (rest[0] != '+' && rest[0] != '-') {
		return dayBound{}, fmt.Errorf("%q is not today, today+N or today-N", arg)
	}
	if days > MaxDays || days < -MaxDays {
		return dayBound{}, fmt.Errorf("%q exceeds the maximum of %d days either way", arg, MaxDays)
	}
	return dayBound{days: days, relative: true}, nil
}

var (
	firstDay = time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
	lastDay  = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)
)

// on is the bound's day, with today the caller's date. A relative bound stops at year 0000 or 9999,
// the years a date prints in four digits, and the clamp keeps the order of two bounds.
func (b dayBound) on(today time.Time) time.Time {
	if !b.relative {
		return b.day
	}
	year, month, day := today.Date()
	return clampDay(time.Date(year, month, day+b.days, 0, 0, 0, 0, time.UTC))
}

func (b dayBound) after(other dayBound) bool {
	if b.relative {
		return b.days > other.days
	}
	return b.day.After(other.day)
}

// namedBound is parseDayBound with the error naming which bound of date it was.
func namedBound(which, arg string) (dayBound, error) {
	b, err := parseDayBound(arg)
	switch {
	case err == nil:
		return b, nil
	case strings.HasPrefix(arg, "today"):
		return b, fmt.Errorf("date(from,to,layout): %s %w", which, err)
	}
	return b, fmt.Errorf("date(from,to,layout): %s %q is not a YYYY-MM-DD date, today, today+N or today-N", which, arg)
}

// dayBounds reads the two bounds of a date call, refusing a pair that cannot be ordered at load.
func dayBounds(a []string) (from, to dayBound, err error) {
	if from, err = namedBound("from", a[0]); err != nil {
		return from, to, err
	}
	if to, err = namedBound("to", a[1]); err != nil {
		return from, to, err
	}
	switch {
	case from.relative != to.relative:
		return from, to, fmt.Errorf("date(from,to,layout): from %q and to %q are not both days or both relative to today, as today-30 and today+14", a[0], a[1])
	case from.after(to):
		return from, to, fmt.Errorf("date(from,to,layout): from %s is after to %s", a[0], a[1])
	}
	return from, to, nil
}

func clampDay(d time.Time) time.Time {
	switch {
	case d.Before(firstDay):
		return firstDay
	case d.After(lastDay):
		return lastDay
	}
	return d
}
