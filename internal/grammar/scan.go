package grammar

// scanner walks a token body, skipping each byte inside a [selector] and, while quotes is set, inside
// a 'quote'. A quote opens only outside a selector. A path scans with quotes unset, since it may hold
// an apostrophe, as o'neil does.
type scanner struct{ quotes bool }

// find calls match with each index of s outside a selector and a quote, in order, and returns the
// first index match reports true for, or -1. match may set sc.quotes, so quotes count from where a
// clause opens.
func (sc *scanner) find(s string, match func(i int) bool) int {
	var st scanState
	for i := 0; i < len(s); i++ {
		if !st.inside(s[i], sc.quotes) && match(i) {
			return i
		}
	}
	return -1
}

// scanState is where a scan stands: how many selectors are open, and whether a quote is open.
type scanState struct {
	depth  int
	quoted bool
}

// inside moves past c and reports whether it belongs to a selector or a quote, and so is no place
// for a match; quotes says whether a quote may open here.
func (st *scanState) inside(c byte, quotes bool) bool {
	switch {
	case st.quoted:
		st.quoted = c != '\''
	case st.depth > 0 || c == '[':
		st.selector(c)
	case c == '\'' && quotes:
		st.quoted = true
	default:
		return false
	}
	return true
}

// selector moves past c, a byte inside a selector or the "[" that opens one.
func (st *scanState) selector(c byte) {
	switch c {
	case '[':
		st.depth++
	case ']':
		st.depth--
	}
}
