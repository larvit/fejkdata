package grammar

// scanner walks a token body, skipping each byte inside a [selector] and, while quotes is set, inside
// a 'quote'. A quote opens only outside a selector. A path scans with quotes unset, since it may hold
// an apostrophe, as o'neil does.
type scanner struct{ quotes bool }

// find calls match with each index of s outside a selector and a quote, in order, and returns the
// first index match reports true for, or -1. match may set sc.quotes, so quotes count from where a
// clause opens.
func (sc *scanner) find(s string, match func(i int) bool) int {
	depth, quoted := 0, false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case quoted:
			quoted = c != '\''
		case c == '[':
			depth++
		case c == ']' && depth > 0:
			depth--
		case depth > 0:
		case c == '\'' && sc.quotes:
			quoted = true
		case match(i):
			return i
		}
	}
	return -1
}
