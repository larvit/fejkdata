package fejkdata

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
)

// cprDigits reads a CPR number, written with or without its hyphen, as ten digits.
func cprDigits(s string) (d [10]int, ok bool) {
	if len(s) == 11 && s[6] == '-' {
		s = s[:6] + s[7:]
	}
	if len(s) != 10 {
		return d, false
	}
	for i := range d {
		if s[i] < '0' || s[i] > '9' {
			return d, false
		}
		d[i] = int(s[i] - '0')
	}
	return d, true
}

// validCpr is CPR-kontoret's check in "Personnummeret i CPR-systemet", appendix "Beregning af
// kontrolciffer": the ten digits, weighted 4 3 2 7 6 5 4 3 2 1, add up to a multiple of 11.
func validCpr(s string) bool {
	d, ok := cprDigits(s)
	if !ok {
		return false
	}
	sum := 0
	for i, w := range []int{4, 3, 2, 7, 6, 5, 4, 3, 2, 1} {
		sum += w * d[i]
	}
	return sum%11 == 0
}

// cprYear is the year the same document's table of "Personnummerets opbygning" gives the year
// digits and the seventh digit.
func cprYear(yy, seventh int) int {
	switch {
	case seventh <= 3:
		return 1900 + yy
	case seventh == 4 || seventh == 9:
		if yy <= 36 {
			return 2000 + yy
		}
		return 1900 + yy
	default:
		if yy <= 57 {
			return 2000 + yy
		}
		return 1800 + yy
	}
}

func TestValidCprAcceptsTheWorkedExample(t *testing.T) {
	if !validCpr("070761-4285") || !validCpr("0707614285") {
		t.Error("070761-4285 is CPR-kontoret's worked example, want it valid")
	}
}

func TestValidCprRefusesAWrongCheckDigitLengthOrCharacter(t *testing.T) {
	for _, bad := range []string{"070761-4286", "070761-428", "070761-42851", "07076x-4285"} {
		t.Run(bad, func(t *testing.T) {
			if validCpr(bad) {
				t.Errorf("%s is no CPR number with a valid check digit", bad)
			}
		})
	}
}

func engineOn(seed uint64, day time.Time) *Generator {
	return &Generator{drawState: drawstate.New(seed, day)}
}

var cprLastDigits = strings.Split("0123456789", "")

func eachCprAfter070761(t *testing.T, f *Generator, last string, check func(got string)) {
	t.Helper()
	eachDraw(300, func() string {
		return mustRender(t, f, `{"format":"{pre}-{cpr(last)}","pre":"070761","last":"`+last+`"}`)
	}, check)
}

func TestBuiltinCprDrawsAValidNumberWhoseLastDigitHasTheParityOfItsOperand(t *testing.T) {
	f := emptyGenerator(1)
	for _, last := range cprLastDigits {
		t.Run(last, func(t *testing.T) {
			odd := last[0]%2 == 1
			eachCprAfter070761(t, f, last, func(got string) {
				if len(got) != 11 || got[:7] != "070761-" || !validCpr(got) || (got[10]%2 == 1) != odd {
					t.Fatalf("cpr after 070761 for %q = %q, want a valid number whose last digit has the parity of %s", last, got, last)
				}
			})
		})
	}
}

func TestBuiltinCprDrawsASeventhDigitOf1937To1999(t *testing.T) {
	f := emptyGenerator(1)
	for _, last := range cprLastDigits {
		t.Run(last, func(t *testing.T) {
			eachCprAfter070761(t, f, last, func(got string) {
				if !strings.ContainsRune("012349", rune(got[7])) {
					t.Fatalf("cpr = %q, want a seventh digit of 1937 to 1999", got)
				}
			})
		})
	}
}

func TestBuiltinCprVariesItsDigits(t *testing.T) {
	f := emptyGenerator(1)
	for _, last := range cprLastDigits {
		t.Run(last, func(t *testing.T) {
			seen := tallyDraws(300, func() string {
				return mustRender(t, f, `{"format":"{pre}-{cpr(last)}","pre":"070761","last":"`+last+`"}`)
			})
			if len(seen) < 20 {
				t.Errorf("cpr drew %d numbers in 300 for %q, want its digits to vary", len(seen), last)
			}
		})
	}
}

func cprEighthAndNinthDigitsDrawn(t *testing.T, f *Generator) (eighths, ninths map[byte]bool) {
	t.Helper()
	eighths, ninths = map[byte]bool{}, map[byte]bool{}
	for _, last := range cprLastDigits {
		eachCprAfter070761(t, f, last, func(got string) {
			eighths[got[8]] = true
			ninths[got[9]] = true
		})
	}
	return eighths, ninths
}

func TestBuiltinCprDrawsEveryEighthAndNinthDigit(t *testing.T) {
	eighths, ninths := cprEighthAndNinthDigitsDrawn(t, emptyGenerator(1))
	if len(eighths) != 10 || len(ninths) != 10 {
		t.Errorf("cpr drew the eighth digits %v and ninth digits %v, want all ten of each", eighths, ninths)
	}
}

func TestBuiltinCprReadsItsOperandFromAPath(t *testing.T) {
	f := emptyGenerator(2)
	eachDraw(50, func() string {
		return mustRender(t, f, `{"format":"{pre}{cpr(sex.last)}","pre":"070761","sex":{"format":"{last}","last":"4"}}`)
	}, func(got string) {
		if !validCpr(got) || got[9]%2 != 0 {
			t.Fatalf("cpr = %q, want a valid number ending in an even digit", got)
		}
	})
}

func TestBuiltinCprRendersNothingForAnOperandThatIsNotOneDigit(t *testing.T) {
	f := emptyGenerator(3)
	for _, last := range []string{"", "mand", "12", "x7", "-"} {
		t.Run(fmt.Sprintf("operand %q", last), func(t *testing.T) {
			tmpl := `{"format":"{pre}-{cpr(last)}","pre":"070761","last":"` + last + `"}`
			if got := mustRender(t, f, tmpl); got != "070761-" {
				t.Errorf("cpr with operand %q = %q, want 070761- and nothing after it", last, got)
			}
		})
	}
}

func TestBuiltinCprReadsTheLastSixDigitsEmitted(t *testing.T) {
	f := emptyGenerator(4)
	for pre, six := range map[string]string{"x-070761": "070761", "99070761": "070761", "a0b7c0d7e6f1g": "070761", "061": "000061", "5": "000005", "x": "000000"} {
		t.Run(pre, func(t *testing.T) {
			got := mustRender(t, f, `{"format":"{pre}{cpr(last)}","pre":"`+pre+`","last":"3"}`)
			if tail := got[len(pre):]; !validCpr(six + tail) {
				t.Errorf("cpr after %q = %q, want check digits valid for %s", pre, tail, six)
			}
		})
	}
}

type cprYearCase struct {
	today time.Time
	yy    string
	want  string
}

func (c cprYearCase) name() string { return c.today.Format("2006-01-02") + " 0101" + c.yy }

var cprYearCases = []cprYearCase{
	{testDay, "00", "456789"},
	{testDay, "26", "456789"},
	{testDay, "27", "0123"},
	{testDay, "36", "0123"},
	{testDay, "37", "012349"},
	{testDay, "99", "012349"},
	{time.Date(2040, 6, 1, 0, 0, 0, 0, time.UTC), "39", "5678"},
	{time.Date(2040, 6, 1, 0, 0, 0, 0, time.UTC), "40", "5678"},
	{time.Date(2040, 6, 1, 0, 0, 0, 0, time.UTC), "41", "012349"},
	{time.Date(2040, 6, 1, 0, 0, 0, 0, time.UTC), "03", "456789"},
}

func latestYearBearing(today time.Time, yy string) int {
	year, _ := strconv.Atoi(yy)
	latest := today.Year()
	for latest%100 != year {
		latest--
	}
	return latest
}

func eachCprOfTheYearDigits(t *testing.T, c cprYearCase, check func(got string)) {
	t.Helper()
	f := engineOn(5, c.today)
	eachDraw(600, func() string {
		return mustRender(t, f, `{"format":"{pre}{cpr(last)}","pre":"0101`+c.yy+`","last":"1"}`)
	}, check)
}

// The year digits name the latest year not after today that bears them, so a date drawn from
// the last hundred years has one century, and the seventh digit is one the table gives that year.
func TestBuiltinCprDrawsANumberOfTheYearItsDigitsName(t *testing.T) {
	for _, c := range cprYearCases {
		t.Run(c.name(), func(t *testing.T) {
			year := latestYearBearing(c.today, c.yy)
			eachCprOfTheYearDigits(t, c, func(got string) {
				if !validCpr(got) || cprYear(year%100, int(got[6]-'0')) != year {
					t.Fatalf("on %s, 0101%s gave %q, want a valid number of the year %d", c.today.Format("2006-01-02"), c.yy, got, year)
				}
			})
		})
	}
}

func TestBuiltinCprDrawsEverySeventhDigitTheTableGivesTheYear(t *testing.T) {
	for _, c := range cprYearCases {
		t.Run(c.name(), func(t *testing.T) {
			seen := map[byte]bool{}
			eachCprOfTheYearDigits(t, c, func(got string) { seen[got[6]] = true })
			var digits []byte
			for d := range seen {
				digits = append(digits, d)
			}
			slices.Sort(digits)
			if string(digits) != c.want {
				t.Errorf("on %s, 0101%s drew the seventh digits %s, want %s", c.today.Format("2006-01-02"), c.yy, digits, c.want)
			}
		})
	}
}

func TestBuiltinCprRendersNothingForAYearTheTableCannotWrite(t *testing.T) {
	f := engineOn(6, time.Date(2060, 1, 1, 0, 0, 0, 0, time.UTC))
	if got := mustRender(t, f, `{"format":"010159{cpr(last)}","last":"1"}`); got != "010159" {
		t.Errorf("cpr in 2059 = %q, want 010159 and nothing after it: the table ends at 2057", got)
	}
}
