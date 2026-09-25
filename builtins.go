package fejkdata

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
// samples read only the rng. A time-based id (uuid v7, ulid) draws its timestamp
// from the rng, not the wall clock, so seeded output stays reproducible.
var builtins = withTransforms(map[string]builtin{
	"luhn":   {arity: 0, prep: derive(func(e string) string { return string(rune('0' + luhnCheck(e))) })},
	"mod11":  {arity: 0, prep: derive(mod11Check)},
	"ean":    {arity: 0, prep: derive(eanCheck)},
	"uuid":   {arity: 0, prep: sample(uuidV7)},
	"ulid":   {arity: 0, prep: sample(ulid)},
	"nanoid": {arity: 1, checkArgs: posIntArg, prep: chars(nanoidAlphabet)},
	"hex":    {arity: 1, checkArgs: posIntArg, prep: chars(hexDigits)},
	"digits": {arity: 1, checkArgs: posIntArg, prep: chars("0123456789"), prints: DataTypeString, proveNumber: func(token string, prints DataType, a []string) proven {
		return printing(token, prints, bounded(0, math.Pow(10, float64(atoi(a[0])))-1, true))
	}},
	"upper": {arity: 1, checkArgs: posIntArg, prep: chars("ABCDEFGHIJKLMNOPQRSTUVWXYZ")},
	"lower": {arity: 1, checkArgs: posIntArg, prep: chars("abcdefghijklmnopqrstuvwxyz")},
	"base64": {arity: 1, checkArgs: posIntArg, prep: func(a []string) callFn {
		n := atoi(a[0])
		return func(s *session, _ string, _ []string) string {
			return base64.StdEncoding.EncodeToString(randBytes(s, n))
		}
	}},
	"int": {arity: 2, checkArgs: intRangeArgs, prep: func(a []string) callFn {
		lo, span := atoi(a[0]), atoi(a[1])-atoi(a[0])+1
		return func(s *session, _ string, _ []string) string { return strconv.Itoa(lo + s.IntN(span)) }
	}, prints: DataTypeInteger, proveNumber: func(token string, prints DataType, a []string) proven {
		return printing(token, prints, bounded(float64(atoi(a[0])), float64(atoi(a[1])), true))
	}},
	"float": {arity: 3, checkArgs: floatArgs, prep: func(a []string) callFn {
		lo, hi, dp := atof(a[0]), atof(a[1]), atoi(a[2])
		return func(s *session, _ string, _ []string) string {
			return formatFloat(lo+s.Float64()*(hi-lo), dp)
		}
	}, prints: DataTypeNumber, proveNumber: func(token string, _ DataType, a []string) proven {
		return printedNumber(token, bounded(atof(a[0]), atof(a[1]), false), atoi(a[2]))
	}},
	"iban": {arity: 1, checkArgs: ibanArg, prep: func(a []string) callFn {
		cc := a[0]
		return func(s *session, _ string, _ []string) string { return iban(s, cc) }
	}},
	"date": {arity: -1, checkArgs: dateArgs, prep: datePrep},
	"time": {arity: -1, checkArgs: timeArg, prep: timePrep},
	"calc": {arity: -1, checkArgs: checkCalc, prep: calcPrep, operands: calcOperands},
	// seq is the one stateful builtin: a per-session counter from 1, advancing on
	// each call. An optional name selects an independent counter; no name uses the
	// default one. Deterministic by construction, so seeded output stays stable.
	"seq": {arity: -1, checkArgs: seqArg, prep: func(a []string) callFn {
		key := ""
		if len(a) == 1 {
			key = a[0]
		}
		return func(s *session, _ string, _ []string) string {
			return strconv.FormatUint(s.next(key), 10)
		}
	}, prints: DataTypeInteger, proveNumber: func(token string, prints DataType, _ []string) proven {
		return printing(token, prints, bounded(1, math.MaxInt64, true))
	}},
})

// derive and sample are the two argument-free builtin shapes: a derivation reads
// the output emitted so far, a sample reads only the rng. chars is the shape of a
// sample of n characters drawn from an alphabet.
func derive(f func(emitted string) string) func([]string) callFn {
	return func([]string) callFn {
		return func(_ *session, emitted string, _ []string) string { return f(emitted) }
	}
}
func sample(f func(rng) string) func([]string) callFn {
	return func([]string) callFn {
		return func(s *session, _ string, _ []string) string { return f(s) }
	}
}
func chars(alphabet string) func([]string) callFn {
	return func(a []string) callFn {
		n := atoi(a[0])
		return func(s *session, _ string, _ []string) string { return randChars(s, n, alphabet) }
	}
}

const hexDigits = "0123456789abcdef"

// formatFloat prints v to dp decimals, -1 for the shortest form, and a zero unsigned.
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
		panic(fmt.Sprintf("fejkdata: builtin arg %q reached prep unvalidated: %v", s, err))
	}
	return n
}

// atof is atoi for a float arg, and reports an unvalidated one the same way.
func atof(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Sprintf("fejkdata: builtin arg %q reached prep unvalidated: %v", s, err))
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

// plainInt parses an integer arg written the one way: no sign, no leading zero.
func plainInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if errors.Is(err, strconv.ErrRange) {
		return 0, fmt.Errorf("%q is past the integer range: %w", s, err)
	}
	if err != nil {
		return 0, fmt.Errorf("%q is not an integer", s)
	}
	if strconv.Itoa(n) != s {
		return 0, fmt.Errorf("%q is not a plain integer; write %d", s, n)
	}
	return n, nil
}
func posIntArg(_ map[string]node, a []string) error {
	n, err := plainInt(a[0])
	if errors.Is(err, strconv.ErrRange) {
		return fmt.Errorf("count %q exceeds the maximum %d", a[0], maxLen)
	}
	if err != nil {
		return fmt.Errorf("count %w", err)
	}
	if n < 1 {
		return fmt.Errorf("count %q must be positive", a[0])
	}
	if n > maxLen {
		return fmt.Errorf("count %d exceeds the maximum %d", n, maxLen)
	}
	return nil
}
func intRangeArgs(_ map[string]node, a []string) error {
	lo, err := plainInt(a[0])
	if err != nil {
		return fmt.Errorf("int(min,max): min %w", err)
	}
	hi, err := plainInt(a[1])
	if err != nil {
		return fmt.Errorf("int(min,max): max %w", err)
	}
	if lo > hi {
		return fmt.Errorf("int(min,max): min %d > max %d", lo, hi)
	}
	if lo == hi {
		return fmt.Errorf("int(%d,%d) is the constant %d; write it as text", lo, hi, lo)
	}
	if uint64(hi)-uint64(lo) >= uint64(math.MaxInt64) { // span hi-lo+1 would overflow int -> IntN panic
		return fmt.Errorf("int(min,max): range %d..%d is too wide", lo, hi)
	}
	return nil
}
func floatArgs(_ map[string]node, a []string) error {
	lo, e1 := strconv.ParseFloat(a[0], 64)
	hi, e2 := strconv.ParseFloat(a[1], 64)
	if e1 != nil || e2 != nil {
		return fmt.Errorf("float(min,max,dp) needs numeric bounds, got %q,%q", a[0], a[1])
	}
	dp, err := plainInt(a[2])
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
	if lo == hi {
		return fmt.Errorf("float(%s,%s,%d) is the constant %q; write it as text", a[0], a[1], dp, strconv.FormatFloat(lo, 'f', dp, 64))
	}
	if math.IsInf(hi-lo, 0) { // an overflowing span would render as "+Inf"
		return fmt.Errorf("float(min,max,dp): range %v..%v is too wide", lo, hi)
	}
	return nil
}
func seqArg(_ map[string]node, a []string) error {
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
