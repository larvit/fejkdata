package builtinfunc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// AddDaysName is the builtin whose date operand package fejkdata proves, since only it holds
// the nodes the operand reads.
const AddDaysName = "adddays"

// MaxDays caps adddays' offset, so a date it reads, from year 0 to 9999, stays a time.Time.
const MaxDays = 1000000

func addDaysArgs(a []string) error {
	if len(a) < 2 || len(a) > 3 {
		hint := ""
		if len(a) > 3 && !holdsQuotedLayout(a[2:]) {
			hint = fmt.Sprintf("; a layout holding a comma is quoted: '%s'", strings.Trim(strings.Join(a[2:], ", "), `'"`))
		}
		return fmt.Errorf("adddays takes 2 or 3 arguments, got %d%s", len(a), hint)
	}
	for _, arg := range a[:2] {
		if _, _, isCall := grammar.FuncCall(arg); isCall {
			return errors.New("adddays(d,n) takes a name, a field or a path for d, and for n one of those or an integer")
		}
	}
	if days, literal, err := literalDays(a[1]); literal && err != nil {
		return fmt.Errorf("adddays(d,n): n %w", err)
	} else if literal && (days > MaxDays || days < -MaxDays) {
		return fmt.Errorf("adddays(d,n): n %d exceeds the maximum of %d days either way", days, MaxDays)
	}
	if len(a) == 3 {
		_, err := layoutArg(a[2])
		return err
	}
	return nil
}

// literalDays reads n as an integer where it starts as one does, with a digit or a sign;
// anything else names what holds the count.
func literalDays(n string) (int, bool, error) {
	if n == "" || !strings.ContainsRune("+-0123456789", rune(n[0])) {
		return 0, false, nil
	}
	days, err := intArg(n)
	return days, true, err
}

func addDaysOperands(a []string) []string {
	if _, literal, _ := literalDays(a[1]); literal {
		return a[:1]
	}
	return a[:2]
}

// DateLayout is the layout a date or born call prints in, false for any other call.
func DateLayout(name string, args []string) (string, bool) {
	if name != "date" && name != "born" {
		return "", false
	}
	return layoutOf(args[2]), true
}

// ReadsBack reports whether a date printed in layout parses back as the same day, from year
// 0 to 9999, as a layout with a two-digit year or no month does not.
func ReadsBack(layout string) bool {
	for _, d := range []time.Time{time.Date(1, 2, 3, 4, 5, 6, 0, time.UTC), time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC), time.Date(9999, 11, 28, 0, 0, 0, 0, time.UTC)} {
		back, err := time.Parse(layout, d.Format(layout))
		if err != nil || back.YearDay() != d.YearDay() || back.Year() != d.Year() {
			return false
		}
	}
	return true
}

// AddDays is an adddays call, which Check accepted, reading its date in from, a layout that
// ReadsBack, and printing it in its own layout, or in from where it gives none.
func AddDays(from string, a []string) Call {
	to := from
	if len(a) == 3 {
		to = layoutOf(a[2])
	}
	days, literal, _ := literalDays(a[1])
	return func(_ *drawstate.State, _ string, operands []string) string {
		d, err := time.Parse(from, operands[0])
		if err != nil {
			panic(invariant.Broken("adddays read %q, which its layout %q cannot parse: %v", operands[0], from, err))
		}
		n := days
		if !literal {
			v, err := strconv.ParseFloat(strings.TrimSpace(operands[1]), 64)
			if err != nil {
				panic(invariant.Broken("adddays read %q as its day count, which a load proved a whole number", operands[1]))
			}
			n = int(v)
		}
		return d.AddDate(0, 0, n).UTC().Format(to)
	}
}

func addDaysPrep(a []string) Call {
	return func(*drawstate.State, string, []string) string {
		panic(invariant.Broken("adddays(%s) rendered before its date's layout was proven", strings.Join(a, ",")))
	}
}
