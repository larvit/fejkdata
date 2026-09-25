package fejkdata

import (
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
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
	f := engine(1)
	for _, c := range cases {
		seen := map[string]bool{}
		for i := 0; i < 200; i++ {
			got := mustRender(t, f, c.tmpl)
			if !c.re.MatchString(got) {
				t.Fatalf("%s = %q, want %s", c.name, got, c.re)
			}
			seen[got] = true
		}
		if len(seen) < 190 {
			t.Fatalf("%s not varied: %d uniques in 200", c.name, len(seen))
		}
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
		if a, b := mustRender(t, engine(7), tmpl), mustRender(t, engine(7), tmpl); a != b {
			t.Fatalf("%s not reproducible: %q != %q", tmpl, a, b)
		}
	}
}

func TestBuiltinIntInclusiveRange(t *testing.T) {
	f := engine(1)
	lo, hi := false, false
	for i := 0; i < 1000; i++ {
		n, err := strconv.Atoi(mustRender(t, f, `"{int(3,7)}"`))
		if err != nil || n < 3 || n > 7 {
			t.Fatalf("int(3,7) = %d (err %v), out of range", n, err)
		}
		lo, hi = lo || n == 3, hi || n == 7
	}
	if !lo || !hi {
		t.Fatalf("int(3,7) never hit a bound: lo=%v hi=%v (bounds must be inclusive)", lo, hi)
	}
}

func TestBuiltinFloat(t *testing.T) {
	f, re := engine(1), regexp.MustCompile(`^[12]\.\d{3}$`)
	for i := 0; i < 200; i++ {
		if got := mustRender(t, f, `"{float(1,2,3)}"`); !re.MatchString(got) {
			t.Fatalf("float(1,2,3) = %q, want d.ddd in [1,2]", got)
		}
		if got := mustRender(t, f, `"{float(-1,1,0)}"`); got == "-0" {
			t.Fatalf("float(-1,1,0) = %q, want a zero printed unsigned", got)
		}
	}
}

func TestBuiltinBase64(t *testing.T) {
	got := mustRender(t, engine(1), `"{base64(9)}"`)
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
	f := engine(1)
	for tmpl, want := range cases {
		if got := mustRender(t, f, tmpl); got != want {
			t.Fatalf("%s = %q, want %q", tmpl, got, want)
		}
	}
}

// TestBuiltinSeqPerSession pins seq's contract: a counter from 1, advancing on
// each call, named counters independent, and the whole thing scoped to one
// session so a fresh Generator restarts at 1.
func TestBuiltinSeqPerSession(t *testing.T) {
	f := engine(1)
	for i := 1; i <= 5; i++ {
		if got := mustRender(t, f, `"{seq()}"`); got != strconv.Itoa(i) {
			t.Fatalf("seq call %d = %q, want %d", i, got, i)
		}
	}
	if got := mustRender(t, f, `"{seq(orders)}"`); got != "1" {
		t.Fatalf("named seq(orders) = %q, want its own count from 1", got)
	}
	if got := mustRender(t, f, `"{seq()}"`); got != "6" {
		t.Fatalf("default seq after the named one = %q, want 6 (counters are independent)", got)
	}
	// repeat drives one counter forward across its renders...
	if got := mustRender(t, engine(2), `{"format":"{seq()}","repeat":3,"separator":","}`); got != "1,2,3" {
		t.Fatalf("repeat seq = %q, want 1,2,3", got)
	}
	// ...and a fresh session restarts from 1.
	if got := mustRender(t, engine(2), `"{seq()}"`); got != "1" {
		t.Fatalf("new session seq = %q, want 1", got)
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
		if _, err := compile(parse(t, tmpl)); err == nil {
			t.Errorf("compile(%s) = nil error, want a limit rejection", tmpl)
		}
	}
	// Ordinary values still compile.
	if _, err := compile(parse(t, `"{hex(16)} {int(-5,5)} {float(0,1,3)} {calc(1+1,2)}"`)); err != nil {
		t.Errorf("compile of normal args failed: %v", err)
	}
}

func TestBuiltinIBAN(t *testing.T) {
	wantLen := map[string]int{"SE": 24, "DE": 22, "NO": 15}
	f := engine(1)
	for cc, n := range wantLen {
		tmpl := `"{iban(` + cc + `)}"`
		for i := 0; i < 100; i++ {
			got := mustRender(t, f, tmpl)
			if got[:2] != cc || len(got) != n {
				t.Fatalf("iban(%s) = %q, want %d chars prefixed %s", cc, got, n, cc)
			}
			if !ibanValid(got) {
				t.Fatalf("iban(%s) = %q fails the mod-97 check", cc, got)
			}
		}
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

// TestRegistryShapes pins the builtin contract compileOps relies on: every entry
// supplies prep, and args parsed at compile only behind a check.
func TestRegistryShapes(t *testing.T) {
	for name, b := range builtins {
		if b.prep == nil {
			t.Errorf("builtin %q has no prep: compileOps would call a nil func", name)
		}
		if b.arity != 0 && b.checkArgs == nil {
			t.Errorf("builtin %q parses args in prep with no check", name)
		}
	}
}

// TestArgGuardsPanic pins the guards that report a builtin arg its check should
// have rejected. No data reaches them — checkFunc runs a builtin's check before
// compileOps ever calls prep — so they are exercised directly.
func TestArgGuardsPanic(t *testing.T) {
	for name, call := range map[string]func(){
		"atoi on an unvalidated arg": func() { atoi("nope") },
		"atof on an unvalidated arg": func() { atof("nope") },
	} {
		mustPanic(t, name, call)
	}
}

// mustPanic fails unless call panics, which is what separates a reported
// invariant break from a silently wrong value.
func mustPanic(t *testing.T, name string, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: no panic, want the invariant reported", name)
		}
	}()
	call()
}

func TestClassBuiltinArgs(t *testing.T) {
	for _, bad := range []string{`"{digits(0)}"`, `"{digits(2000000000)}"`, `"{upper(-1)}"`, `"{lower(x)}"`, `"{digits()}"`, `"{upper(1,2)}"`} {
		if _, err := compile(parse(t, bad)); err == nil {
			t.Errorf("compile(%s) = nil error, want the arg rejected", bad)
		}
	}
	for _, ok := range []string{`"{digits(1048576)}"`, `"{upper(1)}"`, `"{lower(26)}"`} {
		if _, err := compile(parse(t, ok)); err != nil {
			t.Errorf("compile(%s) = %v", ok, err)
		}
	}
}

// TestBuiltinDateAndTime pins the two clock-free samples: date draws a second in
// [from 00:00:00, to 23:59:59], both days reachable, and renders it in the quoted Go
// layout, commas and English names included; time draws a second within one day.
func TestBuiltinDateAndTime(t *testing.T) {
	f := engine(1)
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		got := mustRender(t, f, `"{date(1990-01-01,1990-12-31,'2006-01-02')}"`)
		if d, err := time.Parse("2006-01-02", got); err != nil || d.Year() != 1990 {
			t.Fatalf("date = %q, want a 1990 calendar date (err %v)", got, err)
		}
		seen[got] = true
	}
	if len(seen) < 200 {
		t.Fatalf("date drew %d distinct days of 365 in 500, want a uniform spread", len(seen))
	}
	lo, hi := false, false
	for i := 0; i < 200; i++ {
		got := mustRender(t, f, `"{date(2020-02-28,2020-02-29,'2006-01-02')}"`)
		if got != "2020-02-28" && got != "2020-02-29" {
			t.Fatalf("date(2020-02-28,2020-02-29) = %q, out of range", got)
		}
		lo, hi = lo || got == "2020-02-28", hi || got == "2020-02-29"
	}
	if !lo || !hi {
		t.Fatalf("date never hit a bound: lo=%v hi=%v (bounds must be inclusive)", lo, hi)
	}
	if got := mustRender(t, f, `"{date(2020-07-04,2020-07-05,'January 2, 2006')}"`); got != "July 4, 2020" && got != "July 5, 2020" {
		t.Fatalf("date with a comma in its layout = %q", got)
	}
	rfc := regexp.MustCompile(`^2021-\d\d-\d\dT\d\d:\d\d:\d\dZ$`)
	clock := regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
	ampm := regexp.MustCompile(`^(1[0-2]|[1-9]):[0-5]\d (AM|PM)$`)
	seconds := map[string]bool{}
	for i := 0; i < 300; i++ {
		got := mustRender(t, f, `"{date(2021-01-01,2021-12-31,'2006-01-02T15:04:05Z07:00')}"`)
		if !rfc.MatchString(got) {
			t.Fatalf("date in an RFC 3339 layout = %q, want %s", got, rfc)
		}
		seconds[got[17:19]] = true
		if got := mustRender(t, f, `"{time('15:04')}"`); !clock.MatchString(got) {
			t.Fatalf("time('15:04') = %q, want %s", got, clock)
		}
		if got := mustRender(t, f, `"{time('3:04 PM')}"`); !ampm.MatchString(got) {
			t.Fatalf("time('3:04 PM') = %q, want %s", got, ampm)
		}
		got = mustRender(t, f, `"{date(1950-01-01,2000-12-31,'060102')}-238{luhn()}"`)
		if d := digitsOnly(got); len(d) != 10 || !luhnValid(d) {
			t.Fatalf("a personnummer over date() = %q, want ten Luhn-valid digits", got)
		}
	}
	if len(seconds) < 30 {
		t.Fatalf("date drew %d distinct seconds in 300, want the whole day, not midnight", len(seconds))
	}
}

// TestBuiltinDateArgs pins the New-time checks: bounds are calendar dates in order,
// the layout is quoted, names a field, and for time names no date field.
func TestBuiltinDateArgs(t *testing.T) {
	for tmpl, want := range map[string]string{
		`"{date(1990-13-01,1990-12-31,'2006-01-02')}"`:        "1990-13-01",
		`"{date(1990-12-31,1990-01-01,'2006-01-02')}"`:        "is after",
		`"{date(1990-01-01,1990-01-01,'2006-01-02')}"`:        "write it as text",
		`"{date(1990-01-01,1990-12-31,2006-01-02)}"`:          "'2006-01-02'",
		`"{date(1990-01-01,1990-12-31,'January 2, 2006)}"`:    "'",
		`"{date(1990-01-01,1990-12-31,\"January 2, 2006\")}"`: "quoted: 'January 2, 2006'",
		`"{date(1990-01-01,1990-12-31,'x')}"`:                 "text",
		`"{date(1990-01-01,1990-12-31,'')}"`:                  "text",
		`"{date(1990-01-01,1990-12-31)}"`:                     "3 arguments",
		`"{date(1990-01-01,1990-12-31,January 2, 2006)}"`:     "'January 2, 2006'",
		`"{time(3:04 PM, Mon)}"`:                              "'3:04 PM, Mon'",
		`"{time(15:04)}"`:                                     "'15:04'",
		`"{time('2006-01-02 15:04')}"`:                        "date(",
		`"{time('x')}"`:                                       "text",
		`"{date(1990-01-01,1990-12-31,'15:04')}"`:             "time('15:04')",
	} {
		_, err := compile(parse(t, tmpl))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error mentioning %q", tmpl, err, want)
		}
	}
}

// TestBuiltinLayoutErrorsNameARunnableSpelling pins the two layout errors a user
// can follow: a double-quoted layout is named single-quoted, without its own
// quotes carried into the suggestion, whether it split on a comma or not.
func TestBuiltinLayoutErrorsNameARunnableSpelling(t *testing.T) {
	_, err := compile(parse(t, `"{date(1990-01-01,1990-12-31,\"2006-01-02\")}"`))
	if err == nil || !strings.Contains(err.Error(), "write '2006-01-02'") {
		t.Fatalf("a double-quoted layout = %v, want it named single-quoted", err)
	}
	if strings.Contains(err.Error(), `'"`) {
		t.Errorf("%v names a layout that renders its own quotes", err)
	}
	if !strings.Contains(err.Error(), "is double-quoted") {
		t.Errorf("%v does not say which quotes were wrong", err)
	}
	_, err = compile(parse(t, `"{time(15:04)}"`))
	if err == nil || !strings.Contains(err.Error(), "write '15:04'") {
		t.Fatalf("a bare layout = %v, want it named quoted", err)
	}
	// The comma hint belongs to a layout that split, not to a call given extra args.
	_, err = compile(parse(t, `"{time(0,12,'15:04')}"`))
	if err == nil || !strings.Contains(err.Error(), "takes 1 argument, got 3") {
		t.Fatalf("time with three args = %v, want the count named in the singular", err)
	}
	if strings.Contains(err.Error(), "holding a comma") {
		t.Errorf("%v offers the comma hint though the layout is already quoted", err)
	}
}

// TestBuiltinDateSpansOneDay pins that from == to is a day: with a clock layout it
// draws every second of it, and without one it could only emit one value.
func TestBuiltinDateSpansOneDay(t *testing.T) {
	f := engine(1)
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		got := mustRender(t, f, `"{date(2026-01-01,2026-01-01,'2006-01-02 15:04:05')}"`)
		if !strings.HasPrefix(got, "2026-01-01 ") {
			t.Fatalf("date over one day = %q, out of range", got)
		}
		seen[got] = true
	}
	if len(seen) < 400 {
		t.Fatalf("date over one day drew %d distinct seconds in 500, want the whole day", len(seen))
	}
	_, err := compile(parse(t, `"{date(2026-01-01,2026-01-01,'2006-01-02')}"`))
	if err == nil || !strings.Contains(err.Error(), "write it as text") {
		t.Fatalf("one day in a date-only layout = %v, want it named a constant", err)
	}
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
	f := engine(7)
	for tmpl, re := range cases {
		for i := 0; i < 100; i++ {
			if got := mustRender(t, f, tmpl); !re.MatchString(got) {
				t.Fatalf("%s produced %q, want %s", tmpl, got, re)
			}
		}
	}
}

func TestFunctionTokenLuhn(t *testing.T) {
	// {luhn()} appends a Luhn check digit over the digits emitted so far in the
	// current expansion (non-digits skipped but kept). It reads the output
	// buffer, so a value is never re-rendered. Bodies are escaped to fix input.
	f := engine(1)
	cases := map[string]string{
		`"811218987{luhn()}"`:                      "8112189876",  // personnummer body
		`"7992739871{luhn()}"`:                     "79927398713", // classic Luhn vector
		`"811218-987{luhn()}"`:                     "811218-9876", // '-' skipped, kept
		`{"format":"{n}{luhn()}","n":"811218987"}`: "8112189876",  // over a rendered token
	}
	for tmpl, want := range cases {
		if got := mustRender(t, f, tmpl); got != want {
			t.Fatalf("%s = %q, want %q", tmpl, got, want)
		}
	}
}

func TestBuiltinCompileErrors(t *testing.T) {
	for _, bad := range []string{
		`"{digits(0)}"`,        // count must be positive
		`"{upper(x)}"`,         // count must be an integer
		`"{int(a,b)}"`,         // non-integer args
		`"{int(5,1)}"`,         // min > max
		`"{hex(0)}"`,           // count must be positive
		`"{nanoid(-1)}"`,       // negative count
		`"{base64(0)}"`,        // count must be positive
		`"{float(1,2,-1)}"`,    // negative decimals
		`"{float(NaN,NaN,2)}"`, // bounds must be finite
		`"{float(Inf,Inf,2)}"`, // same-sign infinities
		`"{float(1,NaN,2)}"`,   // one NaN bound
		`"{float(-Inf,1,2)}"`,  // one infinite bound
		`"{digits(05)}"`,       // no leading zero
		`"{int(+1,5)}"`,        // a bound is a plain integer
		`"{int(5,5)}"`,         // a constant is written as text
		`"{float(1,1,2)}"`,     // a constant is written as text
		`"{iban(US)}"`,         // unsupported country
		`"{seq(a,b)}"`,         // seq takes at most one name
	} {
		if _, err := compile(parse(t, bad)); err == nil {
			t.Errorf("compile(%s) = nil error, want error", bad)
		}
	}
}

func TestArgErrorsNameTheSpelling(t *testing.T) {
	for src, want := range map[string]string{
		`"{float(1,2,02)}"`:                 "write 2",
		`{"format":"{calc(a,02)}","a":"1"}`: "write 2",
		`"{digits(+5)}"`:                    "write 5",
		`"{hex(99999999999999999999)}"`:     "exceeds the maximum",
		`"{int(007,9)}"`:                    "write 7",
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error saying %q", src, err, want)
		}
	}
}
