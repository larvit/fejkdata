package fejkdata

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

func TestBuiltinIDGenerators(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
		re   *regexp.Regexp
	}{
		{"uuid v7", `"{uuid()}"`, regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)},
		{"ulid", `"{ulid()}"`, regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)},
		{"nanoid", `"{nanoid(21)}"`, regexp.MustCompile(`^[A-Za-z0-9_-]{21}$`)},
		{"hex", `"{hex(16)}"`, regexp.MustCompile(`^[0-9a-f]{16}$`)},
	}
	f := emptyGenerator(1)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			seen := map[string]bool{}
			eachDraw(200, func() string { return mustRender(t, f, c.tmpl) }, func(got string) {
				if !c.re.MatchString(got) {
					t.Fatalf("%s = %q, want %s", c.name, got, c.re)
				}
				seen[got] = true
			})
			if len(seen) < 190 {
				t.Fatalf("%s not varied: %d uniques in 200", c.name, len(seen))
			}
		})
	}
}

// TestBuiltinSamplesReproducible pins the determinism guardrail: every
// sample draws only from the seeded rng, so same seed -> same value.
func TestBuiltinSamplesReproducible(t *testing.T) {
	for _, tmpl := range []string{
		`"{uuid()}"`, `"{ulid()}"`,
		`"{nanoid(12)}"`, `"{int(1,1000000)}"`,
		`"{float(0,1,6)}"`, `"{base64(12)}"`, `"{iban(SE)}"`,
		`"{date(2000-01-01,2020-12-31,'2006-01-02 15:04:05')}"`, `"{time('15:04:05')}"`,
	} {
		t.Run(tmpl, func(t *testing.T) {
			if a, b := mustRender(t, emptyGenerator(7), tmpl), mustRender(t, emptyGenerator(7), tmpl); a != b {
				t.Fatalf("%s not reproducible: %q != %q", tmpl, a, b)
			}
		})
	}
}

func TestBuiltinIntInclusiveRange(t *testing.T) {
	f := emptyGenerator(1)
	lo, hi := false, false
	eachDraw(1000, func() string { return mustRender(t, f, `"{int(3,7)}"`) }, func(got string) {
		n, err := strconv.Atoi(got)
		if err != nil || n < 3 || n > 7 {
			t.Fatalf("int(3,7) = %d (err %v), out of range", n, err)
		}
		lo, hi = lo || n == 3, hi || n == 7
	})
	if !lo || !hi {
		t.Fatalf("int(3,7) never hit a bound: lo=%v hi=%v (bounds must be inclusive)", lo, hi)
	}
}

func TestBuiltinFloat(t *testing.T) {
	f, re := emptyGenerator(1), regexp.MustCompile(`^[12]\.\d{3}$`)
	t.Run("draws within its bounds at its decimals", func(t *testing.T) {
		eachDraw(200, func() string { return mustRender(t, f, `"{float(1,2,3)}"`) }, func(got string) {
			if !re.MatchString(got) {
				t.Fatalf("float(1,2,3) = %q, want d.ddd in [1,2]", got)
			}
		})
	})
	t.Run("prints a zero unsigned", func(t *testing.T) {
		eachDraw(200, func() string { return mustRender(t, f, `"{float(-1,1,0)}"`) }, func(got string) {
			if got == "-0" {
				t.Fatalf("float(-1,1,0) = %q, want a zero printed unsigned", got)
			}
		})
	})
}

func TestBuiltinBase64(t *testing.T) {
	got := mustRender(t, emptyGenerator(1), `"{base64(9)}"`)
	if b, err := base64.StdEncoding.DecodeString(got); err != nil || len(b) != 9 {
		t.Fatalf("base64(9) = %q decodes to %d bytes (err %v), want 9", got, len(b), err)
	}
}

// TestBuiltinChecksums fixes the payload with escapes and compares to a
// hand-computed check, like the luhn test. mod-11 emits X when it would be 10.
func TestBuiltinChecksums(t *testing.T) {
	cases := map[string]string{
		`"12345678{mod11()}"`:   "123456785",     // weights 2..7 from the right
		`"6{mod11()}"`:          "6X",            // remainder 10 -> X
		`"400638133393{ean()}"`: "4006381333931", // EAN-13 (= ISBN-13) check digit
	}
	f := emptyGenerator(1)
	for tmpl, want := range cases {
		t.Run(tmpl, func(t *testing.T) {
			if got := mustRender(t, f, tmpl); got != want {
				t.Fatalf("%s = %q, want %q", tmpl, got, want)
			}
		})
	}
}

func TestBuiltinFnr(t *testing.T) {
	f := emptyGenerator(1)
	seen := map[string]bool{}
	eachDraw(300, func() string {
		return mustRender(t, f, `{"format":"{pre}{fnr(last)}","pre":"0181804","last":"7"}`)
	}, func(got string) {
		if len(got) != 11 || got[:7] != "0181804" || got[8] != '7' || !validFnr(got) {
			t.Fatalf("fnr = %q, want 0181804, an individ digit, 7, and valid check digits", got)
		}
		seen[got] = true
	})
	if len(seen) < 5 {
		t.Errorf("fnr drew %d numbers in 300, want its second individ digit to vary", len(seen))
	}
}

func TestBuiltinFnrReadsItsOperandFromAPath(t *testing.T) {
	f := emptyGenerator(2)
	eachDraw(50, func() string {
		return mustRender(t, f, `{"format":"{pre}{fnr(sex.last)}","pre":"3181804","sex":{"format":"{last}","last":"4"}}`)
	}, func(got string) {
		if !validFnr(got) || got[8] != '4' {
			t.Fatalf("fnr = %q, want a valid number whose individ ends in 4", got)
		}
	})
}

func TestBuiltinFnrRendersNothingForAnOperandThatIsNotOneDigit(t *testing.T) {
	f := emptyGenerator(3)
	for _, last := range []string{"", "mann", "12", "x7", "-"} {
		t.Run(strconv.Quote(last), func(t *testing.T) {
			tmpl := `{"format":"{pre}{fnr(last)}","pre":"0181804","last":"` + last + `"}`
			if got := mustRender(t, f, tmpl); got != "0181804" {
				t.Errorf("fnr with operand %q = %q, want 0181804 and nothing after it", last, got)
			}
		})
	}
}

func TestBuiltinFnrReadsTheLastSevenDigitsEmitted(t *testing.T) {
	f := emptyGenerator(5)
	for pre, seven := range map[string]string{"x-0181804": "0181804", "9990181804": "0181804", "181804": "0181804", "a1b2c3d4e5f6g7h8": "2345678"} {
		t.Run(pre, func(t *testing.T) {
			got := mustRender(t, f, `{"format":"{pre}{fnr(last)}","pre":"`+pre+`","last":"3"}`)
			if tail := got[len(pre):]; !validFnr(seven + tail) {
				t.Errorf("fnr after %q = %q, want check digits valid for %s", pre, tail, seven)
			}
		})
	}
}

func TestValidFnrAcceptsAPublishedTestNumber(t *testing.T) {
	if !validFnr("10108000398") {
		t.Error("10108000398 is a published test number, want it valid")
	}
}

func TestValidFnrRejectsAWrongCheckDigitOrLength(t *testing.T) {
	for _, bad := range []string{"10108000399", "10108000388", "1010800039", "10108000398x"} {
		t.Run(bad, func(t *testing.T) {
			if validFnr(bad) {
				t.Errorf("%s is no valid fødselsnummer", bad)
			}
		})
	}
}

func validFnr(s string) bool {
	if len(s) != 11 {
		return false
	}
	var d [11]int
	for i := range d {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		d[i] = int(s[i] - '0')
	}
	check := func(weights []int, want int) bool {
		sum := 0
		for i, w := range weights {
			sum += w * d[i]
		}
		k := 11 - sum%11
		return k != 10 && k%11 == want
	}
	return check([]int{3, 7, 6, 1, 8, 9, 4, 5, 2}, d[9]) && check([]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}, d[10])
}

// TestBuiltinSeqPerGenerator pins seq's contract: a counter from 1, advancing on
// each call, named counters independent, and the whole thing scoped to one
// draw state so a fresh Generator restarts at 1.
func TestBuiltinSeqPerGenerator(t *testing.T) {
	f := emptyGenerator(1)
	for _, want := range []string{"1", "2", "3", "4", "5"} {
		t.Run("call "+want, func(t *testing.T) {
			if got := mustRender(t, f, `"{seq()}"`); got != want {
				t.Fatalf("seq call %s = %q, want %s", want, got, want)
			}
		})
	}
	t.Run("a named counter counts from 1", func(t *testing.T) {
		if got := mustRender(t, f, `"{seq(orders)}"`); got != "1" {
			t.Fatalf("named seq(orders) = %q, want its own count from 1", got)
		}
	})
	t.Run("the default counter goes on after a named one", func(t *testing.T) {
		if got := mustRender(t, f, `"{seq()}"`); got != "6" {
			t.Fatalf("default seq after the named one = %q, want 6 (counters are independent)", got)
		}
	})
}

func TestBuiltinSeqCountsOnAcrossARepeat(t *testing.T) {
	if got := mustRender(t, emptyGenerator(2), `{"format":"{seq()}","repeat":3,"separator":","}`); got != "1,2,3" {
		t.Fatalf("repeat seq = %q, want 1,2,3", got)
	}
}

func TestBuiltinSeqRestartsInANewGenerator(t *testing.T) {
	if got := mustRender(t, emptyGenerator(2), `"{seq()}"`); got != "1" {
		t.Fatalf("new generator seq = %q, want 1", got)
	}
}

// TestBuiltinArgLimits pins the New-time guards that keep a fat-fingered or
// overflowing argument from panicking or OOM-ing at render. Each must be rejected
// at compile, not detonate when the template is drawn.
func TestBuiltinArgLimits(t *testing.T) {
	bad := []string{
		`"{int(-9223372036854775808,9223372036854775807)}"`, // span overflows int -> IntN panic
		`"{hex(2000000000)}"`,                               // ~2 GB string
		`"{nanoid(2000000000)}"`,
		`"{base64(2000000000)}"`,
		`"{float(-1e308,1e308,2)}"`,                    // span is +Inf
		`"{float(0,1,100000)}"`,                        // decimals -> huge string
		`"{calc(1+1,100000)}"`,                         // calc decimals -> huge string
		`{"format":"{x}","x":"1","repeat":2000000000}`, // repeat -> multi-GB join
	}
	for _, tmpl := range bad {
		t.Run(tmpl, func(t *testing.T) {
			if _, err := resolved(t, tmpl); err == nil {
				t.Errorf("compile(%s) = nil error, want a limit rejection", tmpl)
			}
		})
	}
}

func TestBuiltinArgsWithinTheLimitsCompile(t *testing.T) {
	if _, err := load.Compile(parse(t, `"{hex(16)} {int(-5,5)} {float(0,1,3)} {calc(1+1,2)}"`), tree.NewLoadFinds()); err != nil {
		t.Errorf("compile of normal args failed: %v", err)
	}
}

func TestBuiltinIBAN(t *testing.T) {
	wantLen := map[string]int{"SE": 24, "DE": 22, "NO": 15}
	f := emptyGenerator(1)
	for cc, n := range wantLen {
		t.Run(cc, func(t *testing.T) {
			tmpl := `"{iban(` + cc + `)}"`
			eachDraw(100, func() string { return mustRender(t, f, tmpl) }, func(got string) {
				if got[:2] != cc || len(got) != n {
					t.Fatalf("iban(%s) = %q, want %d chars prefixed %s", cc, got, n, cc)
				}
				if !ibanValid(got) {
					t.Fatalf("iban(%s) = %q fails the mod-97 check", cc, got)
				}
			})
		})
	}
}

// ibanValid checks the IBAN mod-97 rule independently of the builtin: move the
// first four chars to the end, map letters A-Z to 10-35, the number mod 97 == 1.
func ibanValid(s string) bool {
	r := s[4:] + s[:4]
	rem := 0
	for i := 0; i < len(r); i++ {
		switch c := r[i]; {
		case c >= '0' && c <= '9':
			rem = (rem*10 + int(c-'0')) % 97
		case c >= 'A' && c <= 'Z':
			rem = (rem*100 + int(c-'A') + 10) % 97
		default:
			return false
		}
	}
	return rem == 1
}

// mustPanic fails unless call panics with invariant.Broken's phrase, which is what
// separates a reported invariant break from a silently wrong value.
func mustPanic(t *testing.T, name string, call func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil || !strings.HasPrefix(fmt.Sprint(r), invariant.Broken("")) {
			t.Errorf("%s: recovered %v, want the invariant reported as an internal error", name, r)
		}
	}()
	call()
}

func TestClassBuiltinRejectsAnArgThatIsNoCountWithinItsLimit(t *testing.T) {
	for _, bad := range []string{`"{digits(2000000000)}"`, `"{upper(-1)}"`, `"{lower(x)}"`, `"{digits()}"`, `"{upper(1,2)}"`} {
		t.Run(bad, func(t *testing.T) {
			if _, err := resolved(t, bad); err == nil {
				t.Errorf("compile(%s) = nil error, want the arg rejected", bad)
			}
		})
	}
}

func TestClassBuiltinAcceptsACountWithinItsLimit(t *testing.T) {
	for _, ok := range []string{`"{digits(1048576)}"`, `"{upper(1)}"`, `"{lower(26)}"`} {
		t.Run(ok, func(t *testing.T) {
			if _, err := resolved(t, ok); err != nil {
				t.Errorf("compile(%s) = %v", ok, err)
			}
		})
	}
}

// TestBuiltinDateAndTime pins the two clock-free samples: date draws a second in
// [from 00:00:00, to 23:59:59], both days reachable, and renders it in the quoted Go
// layout, commas and English names included; time draws a second within one day.
func TestBuiltinDateAndTime(t *testing.T) {
	f := emptyGenerator(1)
	t.Run("date spreads over its range", func(t *testing.T) {
		seen := map[string]bool{}
		eachDraw(500, func() string { return mustRender(t, f, `"{date(1990-01-01,1990-12-31,'2006-01-02')}"`) }, func(got string) {
			if d, err := time.Parse("2006-01-02", got); err != nil || d.Year() != 1990 {
				t.Fatalf("date = %q, want a 1990 calendar date (err %v)", got, err)
			}
			seen[got] = true
		})
		if len(seen) < 200 {
			t.Fatalf("date drew %d distinct days of 365 in 500, want a uniform spread", len(seen))
		}
	})
	t.Run("date bounds are inclusive", func(t *testing.T) {
		lo, hi := false, false
		eachDraw(200, func() string { return mustRender(t, f, `"{date(2020-02-28,2020-02-29,'2006-01-02')}"`) }, func(got string) {
			if got != "2020-02-28" && got != "2020-02-29" {
				t.Fatalf("date(2020-02-28,2020-02-29) = %q, out of range", got)
			}
			lo, hi = lo || got == "2020-02-28", hi || got == "2020-02-29"
		})
		if !lo || !hi {
			t.Fatalf("date never hit a bound: lo=%v hi=%v (bounds must be inclusive)", lo, hi)
		}
	})
	t.Run("date renders a layout holding a comma", func(t *testing.T) {
		if got := mustRender(t, f, `"{date(2020-07-04,2020-07-05,'January 2, 2006')}"`); got != "July 4, 2020" && got != "July 5, 2020" {
			t.Fatalf("date with a comma in its layout = %q", got)
		}
	})
	t.Run("date draws every second of a day", func(t *testing.T) {
		rfc := regexp.MustCompile(`^2021-\d\d-\d\dT\d\d:\d\d:\d\dZ$`)
		seconds := map[string]bool{}
		eachDraw(300, func() string {
			return mustRender(t, f, `"{date(2021-01-01,2021-12-31,'2006-01-02T15:04:05Z07:00')}"`)
		}, func(got string) {
			if !rfc.MatchString(got) {
				t.Fatalf("date in an RFC 3339 layout = %q, want %s", got, rfc)
			}
			seconds[got[17:19]] = true
		})
		if len(seconds) < 30 {
			t.Fatalf("date drew %d distinct seconds in 300, want the whole day, not midnight", len(seconds))
		}
	})
	t.Run("time renders a 24-hour layout", func(t *testing.T) {
		clock := regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
		eachDraw(300, func() string { return mustRender(t, f, `"{time('15:04')}"`) }, func(got string) {
			if !clock.MatchString(got) {
				t.Fatalf("time('15:04') = %q, want %s", got, clock)
			}
		})
	})
	t.Run("time renders a 12-hour layout", func(t *testing.T) {
		ampm := regexp.MustCompile(`^(1[0-2]|[1-9]):[0-5]\d (AM|PM)$`)
		eachDraw(300, func() string { return mustRender(t, f, `"{time('3:04 PM')}"`) }, func(got string) {
			if !ampm.MatchString(got) {
				t.Fatalf("time('3:04 PM') = %q, want %s", got, ampm)
			}
		})
	})
	t.Run("luhn checks the digits of a date", func(t *testing.T) {
		eachDraw(300, func() string {
			return mustRender(t, f, `"{date(1950-01-01,2000-12-31,'060102')}-238{luhn()}"`)
		}, func(got string) {
			if d := digitsOnly(got); len(d) != 10 || !luhnValid(d) {
				t.Fatalf("a personnummer over date() = %q, want ten Luhn-valid digits", got)
			}
		})
	})
}

// TestBuiltinDateArgs pins the New-time checks: bounds are calendar dates in order, and the
// layout is quoted.
func TestBuiltinDateArgs(t *testing.T) {
	for tmpl, want := range map[string]string{
		`"{date(1990-13-01,1990-12-31,'2006-01-02')}"`:        "1990-13-01",
		`"{time('2006-01-02 15:04')}"`:                        "write date(from,to,layout)",
		`"{date(1990-12-31,1990-01-01,'2006-01-02')}"`:        "is after",
		`"{date(1990-01-01,1990-12-31,2006-01-02)}"`:          "'2006-01-02'",
		`"{date(1990-01-01,1990-12-31,'January 2, 2006)}"`:    "'",
		`"{date(1990-01-01,1990-12-31,\"January 2, 2006\")}"`: "quoted: 'January 2, 2006'",
		`"{date(1990-01-01,1990-12-31)}"`:                     "3 arguments",
		`"{date(1990-01-01,1990-12-31,January 2, 2006)}"`:     "'January 2, 2006'",
		`"{time(3:04 PM, Mon)}"`:                              "'3:04 PM, Mon'",
		`"{time(15:04)}"`:                                     "'15:04'",
	} {
		t.Run(tmpl, func(t *testing.T) {
			_, err := resolved(t, tmpl)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("compile(%s) = %v, want an error mentioning %q", tmpl, err, want)
			}
		})
	}
}

func TestBuiltinLayoutErrorNamesADoubleQuotedLayoutSingleQuoted(t *testing.T) {
	_, err := load.Compile(parse(t, `"{date(1990-01-01,1990-12-31,\"2006-01-02\")}"`), tree.NewLoadFinds())
	if err == nil || !strings.Contains(err.Error(), "write '2006-01-02'") {
		t.Fatalf("a double-quoted layout = %v, want it named single-quoted", err)
	}
	if strings.Contains(err.Error(), `'"`) {
		t.Errorf("%v names a layout that renders its own quotes", err)
	}
	if !strings.Contains(err.Error(), "is double-quoted") {
		t.Errorf("%v does not say which quotes were wrong", err)
	}
}

func TestBuiltinLayoutErrorNamesABareLayoutQuoted(t *testing.T) {
	_, err := load.Compile(parse(t, `"{time(15:04)}"`), tree.NewLoadFinds())
	if err == nil || !strings.Contains(err.Error(), "write '15:04'") {
		t.Fatalf("a bare layout = %v, want it named quoted", err)
	}
}

func TestBuiltinArgCountErrorOffersNoCommaHintForAQuotedLayout(t *testing.T) {
	_, err := load.Compile(parse(t, `"{time(0,12,'15:04')}"`), tree.NewLoadFinds())
	if err == nil || !strings.Contains(err.Error(), "takes 1 argument, got 3") {
		t.Fatalf("time with three args = %v, want the count named in the singular", err)
	}
	if strings.Contains(err.Error(), "holding a comma") {
		t.Errorf("%v offers the comma hint though the layout is already quoted", err)
	}
}

// TestBuiltinDateSpansOneDay pins that from == to is a day: with a clock layout it
// draws every second of it, and without one it prints that day.
func TestBuiltinDateSpansOneDay(t *testing.T) {
	f := emptyGenerator(1)
	t.Run("a clock layout draws the whole day", func(t *testing.T) {
		seen := map[string]bool{}
		eachDraw(500, func() string {
			return mustRender(t, f, `"{date(2026-01-01,2026-01-01,'2006-01-02 15:04:05')}"`)
		}, func(got string) {
			if !strings.HasPrefix(got, "2026-01-01 ") {
				t.Fatalf("date over one day = %q, out of range", got)
			}
			seen[got] = true
		})
		if len(seen) < 400 {
			t.Fatalf("date over one day drew %d distinct seconds in 500, want the whole day", len(seen))
		}
	})
	t.Run("a date-only layout prints that day", func(t *testing.T) {
		if got := mustRender(t, f, `"{date(2026-01-01,2026-01-01,'2006-01-02')}"`); got != "2026-01-01" {
			t.Fatalf("one day in a date-only layout = %q, want that day", got)
		}
	})
}

func TestClassBuiltins(t *testing.T) {
	cases := map[string]*regexp.Regexp{
		`"{digits(1)}"`:          regexp.MustCompile(`^[0-9]$`),
		`"{digits(3)}"`:          regexp.MustCompile(`^[0-9]{3}$`),
		`"{int(1,9)}"`:           regexp.MustCompile(`^[1-9]$`),
		`"{upper(1)}"`:           regexp.MustCompile(`^[A-Z]$`),
		`"{lower(1)}"`:           regexp.MustCompile(`^[a-z]$`),
		`"{upper(2)}{lower(2)}"`: regexp.MustCompile(`^[A-Z]{2}[a-z]{2}$`),
	}
	f := emptyGenerator(7)
	for tmpl, re := range cases {
		t.Run(tmpl, func(t *testing.T) {
			eachDraw(100, func() string { return mustRender(t, f, tmpl) }, func(got string) {
				if !re.MatchString(got) {
					t.Fatalf("%s produced %q, want %s", tmpl, got, re)
				}
			})
		})
	}
}

func TestFunctionTokenLuhn(t *testing.T) {
	// {luhn()} appends a Luhn check digit over the digits emitted so far in the
	// current expansion (non-digits skipped but kept). It reads the output
	// buffer, so a value is never re-rendered. Bodies are escaped to fix input.
	f := emptyGenerator(1)
	cases := map[string]string{
		`"811218987{luhn()}"`:                      "8112189876",  // personnummer body
		`"7992739871{luhn()}"`:                     "79927398713", // classic Luhn vector
		`"811218-987{luhn()}"`:                     "811218-9876", // '-' skipped, kept
		`{"format":"{n}{luhn()}","n":"811218987"}`: "8112189876",  // over a rendered token
	}
	for tmpl, want := range cases {
		t.Run(tmpl, func(t *testing.T) {
			if got := mustRender(t, f, tmpl); got != want {
				t.Fatalf("%s = %q, want %q", tmpl, got, want)
			}
		})
	}
}

func TestBuiltinCompileErrors(t *testing.T) {
	for _, bad := range []string{
		`"{upper(x)}"`,         // count must be an integer
		`"{int(a,b)}"`,         // non-integer args
		`"{int(5,1)}"`,         // min > max
		`"{nanoid(-1)}"`,       // negative count
		`"{float(1,2,-1)}"`,    // negative decimals
		`"{float(NaN,NaN,2)}"`, // bounds must be finite
		`"{float(Inf,Inf,2)}"`, // same-sign infinities
		`"{float(1,NaN,2)}"`,   // one NaN bound
		`"{float(-Inf,1,2)}"`,  // one infinite bound
		`"{iban(US)}"`,         // unsupported country
		`"{seq(a,b)}"`,         // seq takes at most one name
	} {
		t.Run(bad, func(t *testing.T) {
			if _, err := resolved(t, bad); err == nil {
				t.Errorf("compile(%s) = nil error, want error", bad)
			}
		})
	}
}

func TestArgErrorsNameTheSpelling(t *testing.T) {
	if _, err := resolved(t, `"{hex(99999999999999999999)}"`); err == nil || !strings.Contains(err.Error(), "exceeds the maximum") {
		t.Errorf("compile(hex past the range) = %v, want the maximum named", err)
	}
}

func TestASignOrLeadingZeroIsTheInteger(t *testing.T) {
	for long, short := range map[string]string{
		`"{float(1,2,02)}"`:                 `"{float(1,2,2)}"`,
		`{"format":"{calc(a,02)}","a":"1"}`: `{"format":"{calc(a,2)}","a":"1"}`,
		`"{digits(+5)}"`:                    `"{digits(5)}"`,
		`"{digits(05)}"`:                    `"{digits(5)}"`,
		`"{int(007,9)}"`:                    `"{int(7,9)}"`,
		`"{int(+1,5)}"`:                     `"{int(1,5)}"`,
	} {
		t.Run(long, func(t *testing.T) { sameRenders(t, long, short) })
	}
}

func TestBuiltinsShowsEachCall(t *testing.T) {
	got := Builtins()
	for _, name := range []string{"hex(n)", "lower(n)", "digits(n)", "uuid()", "iban(CC)", "seq()", "seq(name)", "date(from,to,'layout')", "slugify(x)"} {
		t.Run(name, func(t *testing.T) {
			if !slices.Contains(got, name) {
				t.Errorf("Builtins() = %v, missing %s", got, name)
			}
		})
	}
}

func TestBuiltinsShowsEachCallOnceInOrder(t *testing.T) {
	got := Builtins()
	if !slices.IsSorted(got) || len(slices.Compact(slices.Clone(got))) != len(got) {
		t.Errorf("Builtins() = %v, want each call once, sorted", got)
	}
}
