package builtinfunc

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/proven"
)

// maxLen caps sample output lengths (hex, nanoid, base64, digits, upper, lower)
// and maxDecimals float/calc decimal places, so a fat-fingered or overflowing
// argument fails at New instead of trying to allocate gigabytes — or panicking —
// at render.
const (
	maxLen      = 1 << 20
	maxDecimals = 1024
)

// builtins is the registry of {name(args)} functions. Derivations read the digits
// emitted so far in the current expansion (place them after their payload);
// samples read only the rng, and born also the date. A time-based id (uuid v7, ulid)
// draws its timestamp from the rng, not the wall clock, so seeded output stays
// reproducible.
var builtins = withTransforms(map[string]builtin{
	"luhn":    {calls: []string{"luhn()", "luhn(n)"}, arity: -1, checkArgs: luhnArg, prep: luhnPrep},
	"mod11":   {calls: []string{"mod11()"}, arity: 0, prep: derive(mod11Check)},
	"ean":     {calls: []string{"ean()"}, arity: 0, prep: derive(eanCheck)},
	"bgmod11": {calls: []string{"bgmod11(x)"}, arity: 1, checkArgs: bgmod11Arg, prep: bgmod11Prep},
	"fnr":     {calls: []string{"fnr(x)"}, arity: 1, checkArgs: fnrArg, prep: fnrPrep, operands: func(a []string) []string { return a }},
	"cpr":     {calls: []string{"cpr(x)"}, arity: 1, checkArgs: cprArg, prep: cprPrep, operands: func(a []string) []string { return a }},
	"uuid":    {calls: []string{"uuid()"}, arity: 0, prep: sample(uuidV7), avoids: fixedRuns(uuidRuns, "00000000-0000-0000-0000-000000000000")},
	"ulid":    {calls: []string{"ulid()"}, arity: 0, prep: sample(ulid), avoids: fixedRuns(ulidRuns, strings.Repeat("0", 26))},
	"nanoid":  {calls: []string{"nanoid(n)"}, arity: 1, checkArgs: countArg, prep: chars(nanoidAlphabet), avoids: charRuns(nanoidAlphabet, '0')},
	"hex":     {calls: []string{"hex(n)"}, arity: 1, checkArgs: countArg, prep: chars(hexDigits), avoids: charRuns(hexDigits, '0')},
	"digits": {calls: []string{"digits(n)"}, arity: 1, checkArgs: countArg, prep: chars(decimalDigits), avoids: charRuns(decimalDigits, '0'), prints: datatype.String, proveNumber: func(token string, prints datatype.DataType, a []string) proven.Facts {
		if atoi(a[0]) == 0 {
			return proven.Unproven("{" + token + "} prints nothing, which is no number")
		}
		return proven.Printing(token, prints, proven.Bounded(0, math.Pow(10, float64(atoi(a[0])))-1, true))
	}},
	"upper": {calls: []string{"upper(n)"}, arity: 1, checkArgs: countArg, prep: chars(upperLetters), avoids: charRuns(upperLetters, 'A')},
	"lower": {calls: []string{"lower(n)"}, arity: 1, checkArgs: countArg, prep: chars(lowerLetters), avoids: charRuns(lowerLetters, 'a')},
	"base64": {calls: []string{"base64(n)"}, arity: 1, checkArgs: countArg, avoids: base64Runs, prep: func(a []string) Call {
		n := atoi(a[0])
		return func(s *drawstate.State, _ string, _ []string) string {
			return base64.StdEncoding.EncodeToString(randBytes(s, n))
		}
	}},
	"int": {calls: []string{"int(min,max)"}, arity: 2, checkArgs: intRangeArgs, prep: func(a []string) Call {
		lo, span := atoi(a[0]), atoi(a[1])-atoi(a[0])+1
		return func(s *drawstate.State, _ string, _ []string) string { return strconv.Itoa(lo + s.IntN(span)) }
	}, prints: datatype.Integer, proveNumber: func(token string, prints datatype.DataType, a []string) proven.Facts {
		return proven.Printing(token, prints, proven.Bounded(float64(atoi(a[0])), float64(atoi(a[1])), true))
	}},
	"float": {calls: []string{"float(min,max,dp)"}, arity: 3, checkArgs: floatArgs, prep: func(a []string) Call {
		lo, hi, dp := atof(a[0]), atof(a[1]), atoi(a[2])
		return func(s *drawstate.State, _ string, _ []string) string {
			return formatFloat(lo+s.Float64()*(hi-lo), dp)
		}
	}, prints: datatype.Number, proveNumber: func(token string, _ datatype.DataType, a []string) proven.Facts {
		return proven.PrintedNumber(token, proven.Bounded(atof(a[0]), atof(a[1]), false), atoi(a[2]))
	}},
	"iban": {calls: []string{"iban(CC)"}, arity: 1, checkArgs: ibanArg, prep: func(a []string) Call {
		cc := a[0]
		return func(s *drawstate.State, _ string, _ []string) string { return iban(s, cc) }
	}},
	"date":      {calls: []string{"date(from,to,'layout')"}, arity: -1, checkArgs: dateArgs, prep: datePrep},
	"born":      {calls: []string{"born(min,max,'layout')"}, arity: -1, checkArgs: bornArgs, prep: bornPrep},
	AddDaysName: {calls: []string{AddDaysName + "(d,n)", AddDaysName + "(d,n,'layout')"}, arity: -1, checkArgs: addDaysArgs, prep: addDaysPrep, operands: addDaysOperands},
	"time":      {calls: []string{"time('layout')"}, arity: -1, checkArgs: timeArg, prep: timePrep},
	RewordName:  {calls: []string{RewordName + "(text,words)"}, arity: 2, checkArgs: pathArgs, prep: renderedByFejkdata, operands: func(a []string) []string { return a }},
	CalcName:    {calls: []string{CalcName + "(expr)", CalcName + "(expr,dp)"}, arity: -1, checkArgs: checkCalc, prep: calcPrep, operands: calcOperands},
	// seq is the one stateful builtin: a per-generator counter from 1, advancing on
	// each call. An optional name selects an independent counter; no name uses the
	// default one. Deterministic by construction, so seeded output stays stable.
	"seq": {calls: []string{"seq()", "seq(name)"}, arity: -1, checkArgs: seqArg, prep: func(a []string) Call {
		key := ""
		if len(a) == 1 {
			key = a[0]
		}
		return func(s *drawstate.State, _ string, _ []string) string {
			return strconv.FormatUint(s.Seq(key), 10)
		}
	}, prints: datatype.Integer, proveNumber: func(token string, prints datatype.DataType, _ []string) proven.Facts {
		return proven.Printing(token, prints, proven.Bounded(1, math.MaxInt64, true))
	}},
})

// derive and sample are the two argument-free builtin shapes: a derivation reads
// the output emitted so far, a sample reads only the rng. chars is the shape of a
// sample of n characters drawn from an alphabet.
func derive(f func(emitted string) string) func([]string) Call {
	return func([]string) Call {
		return func(_ *drawstate.State, emitted string, _ []string) string { return f(emitted) }
	}
}
func sample(f func(rng) string) func([]string) Call {
	return func([]string) Call {
		return func(s *drawstate.State, _ string, _ []string) string { return f(s) }
	}
}
func chars(alphabet string) func([]string) Call {
	return func(a []string) Call {
		n := atoi(a[0])
		return func(s *drawstate.State, _ string, _ []string) string { return randChars(s, n, alphabet) }
	}
}

const (
	hexDigits     = "0123456789abcdef"
	decimalDigits = "0123456789"
	upperLetters  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerLetters  = "abcdefghijklmnopqrstuvwxyz"
)

func charRuns(alphabet string, blank byte) func([]string) (avoid.Runs, string) {
	return func(a []string) (avoid.Runs, string) {
		n := atoi(a[0])
		return avoid.Runs{{Alphabet: alphabet, N: n}}, strings.Repeat(string(blank), n)
	}
}

func fixedRuns(rs avoid.Runs, blank string) func([]string) (avoid.Runs, string) {
	return func([]string) (avoid.Runs, string) { return rs, blank }
}

var (
	uuidRuns = avoid.Runs{{Alphabet: hexDigits, N: 8}, {Alphabet: "-", N: 1}, {Alphabet: hexDigits, N: 4}, {Alphabet: "-", N: 1},
		{Alphabet: "7", N: 1}, {Alphabet: hexDigits, N: 3}, {Alphabet: "-", N: 1}, {Alphabet: "89ab", N: 1}, {Alphabet: hexDigits, N: 3},
		{Alphabet: "-", N: 1}, {Alphabet: hexDigits, N: 12}}
	ulidRuns = avoid.Runs{{Alphabet: "01234567", N: 1}, {Alphabet: crockford, N: 25}}
)

const base64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// base64Runs is the runs of base64(n). Where n is not a multiple of 3, the last character before the
// padding encodes some zero bits, so it is one of 4 or 16 characters.
func base64Runs(a []string) (avoid.Runs, string) {
	n := atoi(a[0])
	rs := avoid.Runs{{Alphabet: base64Alphabet, N: 4 * (n / 3)}}
	switch n % 3 {
	case 1:
		rs = append(rs, avoid.Run{Alphabet: base64Alphabet, N: 1}, avoid.Run{Alphabet: "AQgw", N: 1}, avoid.Run{Alphabet: "=", N: 2})
	case 2:
		rs = append(rs, avoid.Run{Alphabet: base64Alphabet, N: 2}, avoid.Run{Alphabet: "AEIMQUYcgkosw048", N: 1}, avoid.Run{Alphabet: "=", N: 1})
	}
	return rs, base64.StdEncoding.EncodeToString(make([]byte, n))
}

// formatFloat prints v to dp decimals, or proven.ShortestDecimals, and a zero unsigned.
func formatFloat(v float64, dp int) string {
	s := strconv.FormatFloat(v, 'f', dp, 64)
	if strings.HasPrefix(s, "-") && strings.Trim(s, "-0.") == "" {
		return s[1:]
	}
	return s
}

// atoi parses an arg a builtin's check already validated. It panics rather than
// returning zero, so a check that stops covering its own args is a stack trace and
// not a silently wrong length, range or decimal count.
func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(invariant.Broken("builtin arg %q reached prep unvalidated: %v", s, err))
	}
	return n
}

// atof is atoi for a float arg, and reports an unvalidated one the same way.
func atof(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(invariant.Broken("builtin arg %q reached prep unvalidated: %v", s, err))
	}
	return f
}
func randBytes(r rng, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.IntN(256))
	}
	return b
}
func randChars(r rng, n int, alphabet string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[r.IntN(len(alphabet))]
	}
	return string(b)
}

func intArg(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if errors.Is(err, strconv.ErrRange) {
		return 0, fmt.Errorf("%q is past the integer range: %w", s, err)
	}
	if err != nil {
		return 0, fmt.Errorf("%q is not an integer", s)
	}
	return n, nil
}
func countArg(a []string) error {
	n, err := intArg(a[0])
	if errors.Is(err, strconv.ErrRange) {
		return fmt.Errorf("count %q exceeds the maximum %d", a[0], maxLen)
	}
	if err != nil {
		return fmt.Errorf("count %w", err)
	}
	if n < 0 {
		return fmt.Errorf("count %q must not be negative", a[0])
	}
	if n > maxLen {
		return fmt.Errorf("count %d exceeds the maximum %d", n, maxLen)
	}
	return nil
}
func intRangeArgs(a []string) error {
	lo, err := intArg(a[0])
	if err != nil {
		return fmt.Errorf("int(min,max): min %w", err)
	}
	hi, err := intArg(a[1])
	if err != nil {
		return fmt.Errorf("int(min,max): max %w", err)
	}
	if lo > hi {
		return fmt.Errorf("int(min,max): min %d > max %d", lo, hi)
	}
	if uint64(hi)-uint64(lo) >= uint64(math.MaxInt64) { // span hi-lo+1 would overflow int -> IntN panic
		return fmt.Errorf("int(min,max): range %d..%d is too wide", lo, hi)
	}
	return nil
}
func floatArgs(a []string) error {
	lo, e1 := strconv.ParseFloat(a[0], 64)
	hi, e2 := strconv.ParseFloat(a[1], 64)
	if e1 != nil || e2 != nil {
		return fmt.Errorf("float(min,max,dp) needs numeric bounds, got %q,%q", a[0], a[1])
	}
	dp, err := intArg(a[2])
	if err != nil {
		return fmt.Errorf("float(min,max,dp): decimals %w", err)
	}
	if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
		return fmt.Errorf("float(min,max,dp) needs finite bounds, got %q,%q", a[0], a[1])
	}
	if lo > hi {
		return fmt.Errorf("float(min,max,dp): min %v > max %v", lo, hi)
	}
	if dp < 0 || dp > maxDecimals {
		return fmt.Errorf("float(min,max,dp): decimals %d out of range 0..%d", dp, maxDecimals)
	}
	if math.IsInf(hi-lo, 0) { // an overflowing span would render as "+Inf"
		return fmt.Errorf("float(min,max,dp): range %v..%v is too wide", lo, hi)
	}
	return nil
}
func seqArg(a []string) error {
	if len(a) > 1 {
		return fmt.Errorf("seq takes at most one name, got %d", len(a))
	}
	if len(a) == 1 && a[0] == "" {
		return fmt.Errorf("seq name must not be empty")
	}
	return nil
}

// nanoidAlphabet is the 64-char URL-safe set Nano IDs use.
const nanoidAlphabet = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// uuidV7 builds an RFC 9562 v7 UUID. The 48-bit timestamp field is drawn from
// the rng (not the clock) to stay reproducible, then the version (7) and variant
// (10) bits are forced; the rest is random.
func uuidV7(r rng) string {
	b := randBytes(r, 16)
	b[6] = b[6]&0x0f | 0x70
	b[8] = b[8]&0x3f | 0x80
	var sb strings.Builder
	for i, x := range b {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			sb.WriteByte('-')
		}
		sb.WriteByte(hexDigits[x>>4])
		sb.WriteByte(hexDigits[x&0x0f])
	}
	return sb.String()
}

// crockford is the ULID/Crockford base32 alphabet (no I, L, O, U).
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ulid builds a 26-char ULID: 128 random bits (timestamp from the rng) encoded
// big-endian into base32, the 130-bit stream left-padded with two zero bits.
func ulid(r rng) string {
	b := randBytes(r, 16)
	out := make([]byte, 26)
	for i := range out {
		v := 0
		for j := 0; j < 5; j++ { // 5 bits per char; the two leading pad bits are zero
			real := 5*i + j - 2
			bit := 0
			if real >= 0 {
				bit = int(b[real/8]>>(7-uint(real%8))) & 1
			}
			v = v<<1 | bit
		}
		out[i] = crockford[v]
	}
	return string(out)
}
