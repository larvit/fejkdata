package fejkdata

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// maxLen caps sample output lengths (hex, nanoid, base64) and the renders a repeat
// multiplies to along any path; maxDecimals caps float/calc decimal places. So a
// fat-fingered or overflowing argument fails at New instead of trying to allocate
// gigabytes — or panicking — at render.
const (
	maxLen      = 1 << 20 // 1,048,576 chars/bytes
	maxDecimals = 1024
)

// builtins is the registry of {name(args)} functions. Derivations read the digits
// emitted so far in the current expansion (place them after their payload);
// samples read only the rng. A time-based id (uuid v7, ulid) draws its timestamp
// from the rng, not the wall clock, so seeded output stays reproducible.
var builtins = map[string]builtin{
	"luhn":   {arity: 0, prep: derive(func(e string) string { return string(rune('0' + luhnCheck(e))) })},
	"mod11":  {arity: 0, prep: derive(mod11Check)},
	"ean":    {arity: 0, prep: derive(eanCheck)},
	"uuid":   {arity: 0, prep: sample(uuidV7)},
	"ulid":   {arity: 0, prep: sample(ulid)},
	"nanoid": {arity: 1, check: posIntArg, prep: chars(nanoidAlphabet)},
	"hex":    {arity: 1, check: posIntArg, prep: chars(hexDigits)},
	"digits": {arity: 1, check: posIntArg, prep: chars("0123456789")},
	"upper":  {arity: 1, check: posIntArg, prep: chars("ABCDEFGHIJKLMNOPQRSTUVWXYZ")},
	"lower":  {arity: 1, check: posIntArg, prep: chars("abcdefghijklmnopqrstuvwxyz")},
	"base64": {arity: 1, check: posIntArg, prep: func(a []string) callFn {
		n := atoi(a[0])
		return func(s *session, _ string, _ []string) string {
			return base64.StdEncoding.EncodeToString(randBytes(s, n))
		}
	}},
	"int": {arity: 2, check: intRangeArgs, prep: func(a []string) callFn {
		lo, span := atoi(a[0]), atoi(a[1])-atoi(a[0])+1
		return func(s *session, _ string, _ []string) string { return strconv.Itoa(lo + s.IntN(span)) }
	}},
	"float": {arity: 3, check: floatArgs, prep: func(a []string) callFn {
		lo, hi, dp := atof(a[0]), atof(a[1]), atoi(a[2])
		return func(s *session, _ string, _ []string) string {
			return strconv.FormatFloat(lo+s.Float64()*(hi-lo), 'f', dp, 64)
		}
	}},
	"iban": {arity: 1, check: ibanArg, prep: func(a []string) callFn {
		cc := a[0]
		return func(s *session, _ string, _ []string) string { return iban(s, cc) }
	}},
	"calc":      {arity: -1, check: checkCalc, prep: calcPrep, operands: calcOperands},
	"lowercase": {arity: 1, check: transformArg, prep: transformPrep(strings.ToLower), operands: transformOperand},
	"uppercase": {arity: 1, check: transformArg, prep: transformPrep(strings.ToUpper), operands: transformOperand},
	"ascii":     {arity: 1, check: transformArg, prep: transformPrep(asciiFold), operands: transformOperand},
	// seq is the one stateful builtin: a per-session counter from 1, advancing on
	// each call. An optional name selects an independent counter; no name uses the
	// default one. Deterministic by construction, so seeded output stays stable.
	"seq": {arity: -1, check: seqArg, prep: func(a []string) callFn {
		key := ""
		if len(a) == 1 {
			key = a[0]
		}
		return func(s *session, _ string, _ []string) string {
			return strconv.FormatUint(s.next(key), 10)
		}
	}},
}

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

// transforms are the builtins that rewrite one operand's value; they nest, so
// {lowercase(ascii(x))} folds then lowers.
var transforms = map[string]func(string) string{
	"ascii":     asciiFold,
	"lowercase": strings.ToLower,
	"uppercase": strings.ToUpper,
}

// unwrapTransform peels nested transform calls off an operand arg, returning the
// field it finally names and the transforms to apply, innermost last.
func unwrapTransform(arg string) (leaf string, chain []func(string) string, err error) {
	for {
		name, args, isCall := funcCall(arg)
		if !isCall {
			return arg, chain, nil
		}
		fn, isTransform := transforms[name]
		if !isTransform {
			return "", nil, fmt.Errorf("%s(%s) is not a transform, so it cannot be an operand", name, strings.Join(args, ","))
		}
		if len(args) != 1 {
			return "", nil, fmt.Errorf("%s takes 1 arg, got %d", name, len(args))
		}
		chain = append(chain, fn)
		arg = args[0]
	}
}

func transformArg(fields map[string]node, a []string) error {
	leaf, _, err := unwrapTransform(a[0])
	if err != nil {
		return err
	}
	if isRef(leaf) {
		_, _, err := refShape(leaf)
		return err
	}
	return checkArm(leaf, fields)
}

func transformOperand(a []string) []string {
	leaf, _, err := unwrapTransform(a[0])
	if err != nil {
		return nil
	}
	return []string{leaf}
}

func transformPrep(outer func(string) string) func([]string) callFn {
	return func(a []string) callFn {
		_, chain, err := unwrapTransform(a[0])
		if err != nil {
			panic(fmt.Sprintf("fejkdata: transform arg %q reached prep unvalidated: %v", a[0], err))
		}
		return func(_ *session, _ string, operands []string) string {
			v := operands[0]
			for i := len(chain) - 1; i >= 0; i-- {
				v = chain[i](v)
			}
			return outer(v)
		}
	}
}

// asciiFolds maps the Latin letters with diacritics or ligatures to ASCII.
var asciiFolds = map[rune]string{
	'À': "A", 'Á': "A", 'Â': "A", 'Ã': "A", 'Ä': "A", 'Å': "A", 'Æ': "AE", 'Ç': "C",
	'È': "E", 'É': "E", 'Ê': "E", 'Ë': "E", 'Ì': "I", 'Í': "I", 'Î': "I", 'Ï': "I",
	'Ð': "D", 'Ñ': "N", 'Ò': "O", 'Ó': "O", 'Ô': "O", 'Õ': "O", 'Ö': "O", 'Ø': "O",
	'Ù': "U", 'Ú': "U", 'Û': "U", 'Ü': "U", 'Ý': "Y", 'Þ': "Th", 'ß': "ss", 'Œ': "OE",
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a", 'æ': "ae", 'ç': "c",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e", 'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'ð': "d", 'ñ': "n", 'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u", 'ý': "y", 'þ': "th", 'ÿ': "y", 'œ': "oe",
}

// asciiFold rewrites s to ASCII: folded Latin letters stay, any other non-ASCII
// rune is dropped.
func asciiFold(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < unicode.MaxASCII {
			b.WriteRune(r)
		} else {
			b.WriteString(asciiFolds[r])
		}
	}
	return b.String()
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

func posIntArg(_ map[string]node, a []string) error {
	n, err := strconv.Atoi(a[0])
	if err != nil || n < 1 {
		return fmt.Errorf("count %q must be a positive integer", a[0])
	}
	if n > maxLen {
		return fmt.Errorf("count %d exceeds the maximum %d", n, maxLen)
	}
	return nil
}

func intRangeArgs(_ map[string]node, a []string) error {
	lo, e1 := strconv.Atoi(a[0])
	hi, e2 := strconv.Atoi(a[1])
	if e1 != nil || e2 != nil {
		return fmt.Errorf("int(min,max) needs integer args, got %q,%q", a[0], a[1])
	}
	if lo > hi {
		return fmt.Errorf("int(min,max): min %d > max %d", lo, hi)
	}
	if uint64(hi)-uint64(lo) >= uint64(math.MaxInt64) { // span hi-lo+1 would overflow int -> IntN panic
		return fmt.Errorf("int(min,max): range %d..%d is too wide", lo, hi)
	}
	return nil
}

func floatArgs(_ map[string]node, a []string) error {
	lo, e1 := strconv.ParseFloat(a[0], 64)
	hi, e2 := strconv.ParseFloat(a[1], 64)
	dp, e3 := strconv.Atoi(a[2])
	if e1 != nil || e2 != nil || e3 != nil {
		return fmt.Errorf("float(min,max,dp) needs numeric args, got %q,%q,%q", a[0], a[1], a[2])
	}
	if lo > hi {
		return fmt.Errorf("float(min,max,dp): min %v > max %v", lo, hi)
	}
	if math.IsInf(hi-lo, 0) { // an overflowing span would render as "+Inf"
		return fmt.Errorf("float(min,max,dp): range %v..%v is too wide", lo, hi)
	}
	if dp < 0 || dp > maxDecimals {
		return fmt.Errorf("float(min,max,dp): decimals %d out of range 0..%d", dp, maxDecimals)
	}
	return nil
}

func seqArg(_ map[string]node, a []string) error {
	if len(a) > 1 {
		return fmt.Errorf("seq takes at most one name, got %d args", len(a))
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

// luhnCheck returns the Luhn check digit (0-9) over the digits of s; non-digit
// runes are skipped. Doubling runs from the rightmost digit, so the result is
// correct whatever the payload length.
func luhnCheck(s string) int {
	sum, double := 0, true
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		d := int(c - '0')
		if double {
			if d *= 2; d > 9 {
				d -= 9
			}
		}
		double = !double
		sum += d
	}
	return (10 - sum%10) % 10
}

// mod11Check returns the weighted mod-11 check character over the digits of s
// (weights 2..7 cycling from the right). A would-be value of 10 emits 'X', as in
// ISBN-10 / ISO 7064; non-digits are skipped.
func mod11Check(s string) string {
	sum, w := 0, 2
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		sum += int(c-'0') * w
		if w++; w > 7 {
			w = 2
		}
	}
	if chk := (11 - sum%11) % 11; chk != 10 {
		return string(rune('0' + chk))
	}
	return "X"
}

// eanCheck returns the EAN-13 / UPC-A / ISBN-13 / GTIN check digit over the
// digits of s: weights 3 and 1 alternating from the rightmost digit, mod 10.
func eanCheck(s string) string {
	sum, w := 0, 3
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		sum += int(c-'0') * w
		w = 4 - w // 3 <-> 1
	}
	return string(rune('0' + (10-sum%10)%10))
}

// ibanLen maps a supported country code to the full IBAN length. The check digits
// sit between the country code and the BBAN, so — unlike luhn/ean — iban can't be
// a left-to-right derivation; it generates the whole value instead.
var ibanLen = map[string]int{"BE": 16, "DE": 22, "DK": 18, "ES": 24, "FI": 18, "NO": 15, "SE": 24}

func ibanArg(_ map[string]node, a []string) error {
	if _, ok := ibanLen[a[0]]; !ok {
		return fmt.Errorf("iban(%q): unsupported country code", a[0])
	}
	return nil
}

// iban generates a structurally valid IBAN for cc: a numeric BBAN of the right
// length, then mod-97 check digits. Real bank/branch structure isn't modelled —
// the result passes length and checksum validation, which is what fake data needs.
func iban(r rng, cc string) string {
	bban := make([]byte, ibanLen[cc]-4)
	for i := range bban {
		bban[i] = byte('0' + r.IntN(10))
	}
	rem := 0
	feed := func(d int) { rem = (rem*10 + d) % 97 }
	for _, c := range bban {
		feed(int(c - '0'))
	}
	for i := 0; i < len(cc); i++ { // letters A-Z -> 10..35, fed as two digits
		v := int(cc[i]-'A') + 10
		feed(v / 10)
		feed(v % 10)
	}
	feed(0)
	feed(0)
	return fmt.Sprintf("%s%02d%s", cc, 98-rem, bban)
}
