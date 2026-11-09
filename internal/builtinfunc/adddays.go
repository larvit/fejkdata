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

// AddDaysName is the builtin whose date operand package fejkdata proves, since only that package
// holds the nodes the operand reads.
const AddDaysName = "adddays"

// MaxDays caps adddays' offset, so a date it reads, from year 1 to 9999, stays a time.Time.
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
	if err := dayCountArg(a[1]); err != nil {
		return err
	}
	if len(a) == 3 {
		_, err := layoutArg(a[2])
		return err
	}
	return nil
}

func dayCountArg(n string) error {
	days, literal, err := literalDays(n)
	switch {
	case !literal:
		return nil
	case err != nil && len(n) > 1 && strings.ContainsRune("+-", rune(n[0])) && !strings.ContainsRune("0123456789", rune(n[1])):
		return fmt.Errorf("adddays(d,n): n %q signs a name, which adddays cannot read; draw it negative instead, such as {int(-30,-1)}", n)
	case err != nil:
		return fmt.Errorf("adddays(d,n): n %w; write a whole number of days", err)
	case days > MaxDays || days < -MaxDays:
		return fmt.Errorf("adddays(d,n): n %d exceeds the maximum of %d days either way", days, MaxDays)
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

// ReadsBack reports whether a date printed in layout parses back as the same day: every day of a
// leap year at stepped times, and days in years 1, 1969 and 9999. A layout with a two-digit year,
// no day, or fields packed so that one reading is ambiguous, such as '2006-1_2', does not.
func ReadsBack(layout string) bool {
	days := []time.Time{time.Date(1, 2, 3, 4, 5, 6, 0, time.UTC), time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC), time.Date(9999, 11, 28, 0, 0, 0, 0, time.UTC)}
	for d := time.Date(2000, 1, 1, 23, 59, 59, 0, time.UTC); d.Year() == 2000; d = d.Add(24*time.Hour - 11*time.Minute - 7*time.Second) {
		days = append(days, d)
	}
	for _, d := range days {
		back, err := time.Parse(layout, d.Format(layout))
		if err != nil || back.YearDay() != d.YearDay() || back.Year() != d.Year() {
			return false
		}
	}
	return true
}

// AddDays is an adddays call that Check accepted. It parses its date with the layout from, which
// ReadsBack, and prints the result in the call's third argument, or in from when the call has two.
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
