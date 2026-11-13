package builtinfunc

import (
	"fmt"
	"strings"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
)

const (
	timestampLayout = "2006-01-02T15:04:05Z"
	maxTickStep     = 86400
)

// parseTickStart reads a UTC timestamp, or a day as its midnight.
func parseTickStart(arg string) (time.Time, error) {
	if at, err := time.Parse(timestampLayout, arg); err == nil {
		return at, nil
	}
	return time.Parse(dayLayout, arg)
}

func tickArgs(a []string) error {
	if err := layoutArity("tick", 4, a); err != nil {
		return err
	}
	if _, err := parseTickStart(a[0]); err != nil {
		return fmt.Errorf("tick(start,step,jitter,layout): start %q is not a YYYY-MM-DD day or a UTC timestamp such as 2026-01-01T00:00:00Z", a[0])
	}
	step, err := intArg(a[1])
	if err != nil || step < 1 || step > maxTickStep {
		return fmt.Errorf("tick(start,step,jitter,layout): step %q is not a whole number of seconds from 1 to %d", a[1], maxTickStep)
	}
	jitter, err := intArg(a[2])
	if err != nil || jitter < 0 {
		return fmt.Errorf("tick(start,step,jitter,layout): jitter %q is not a whole number of seconds, 0 or more", a[2])
	}
	if jitter >= step {
		return fmt.Errorf("tick(start,step,jitter,layout): jitter %d is not below step %d, so rows could swap order", jitter, step)
	}
	_, err = layoutArg(a[3])
	return err
}

// tickPrep prints row n at start + (n-1) steps + a draw of 0 to jitter seconds. The rows are
// counted by the call's arguments, so equal calls step one counter and others step apart.
func tickPrep(a []string) Call {
	start, err := parseTickStart(a[0])
	if err != nil {
		panic(invariant.Broken("builtin arg %q reached prep unvalidated: %v", a[0], err))
	}
	step, jitter, layout := int64(atoi(a[1])), atoi(a[2]), layoutOf(a[3])
	counter := "tick(" + strings.Join(a, ",") + ")"
	return func(s *drawstate.State, _ string, _ []string) string {
		offset := int64(s.Seq(counter)-1)*step + int64(s.IntN(jitter+1))
		return time.Unix(start.Unix()+offset, int64(start.Nanosecond())).UTC().Format(layout)
	}
}
