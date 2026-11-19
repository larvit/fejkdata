package builtinfunc

import (
	"fmt"
	"strconv"

	"github.com/larvit/fejkdata/internal/drawstate"
)

// nirLength is the characters of a NIR before its key: the sex, the year and month of birth, the
// place of birth and the order number.
const nirLength = 13

// nirPrep derives the key of the NIR whose last nirLength letters and digits were emitted,
// as INSEE defines it: 97 less the remainder of the number divided by 97, where Corsica's 2A and
// 2B count as 19 and 18. It renders nothing where those characters are no NIR.
func nirPrep([]string, parsed) Call {
	return func(_ *drawstate.State, emitted string, _ []string) string {
		n, ok := nirNumber(alnumTail(emitted, nirLength))
		if !ok {
			return ""
		}
		return fmt.Sprintf("%02d", 97-n%97)
	}
}

// alnumTail is the last n ASCII letters and digits of s, fewer where s holds fewer.
func alnumTail(s string, n int) string {
	tail := make([]byte, 0, n)
	for i := len(s) - 1; i >= 0 && len(tail) < n; i-- {
		if c := s[i]; c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			tail = append(tail, c)
		}
	}
	for l, r := 0, len(tail)-1; l < r; l, r = l+1, r-1 {
		tail[l], tail[r] = tail[r], tail[l]
	}
	return string(tail)
}

// nirNumber reads the thirteen characters of a NIR as a number. Only its department, the sixth
// and seventh, may be 2A or 2B, which INSEE counts as 19 and 18.
func nirNumber(s string) (n int, ok bool) {
	if len(s) != nirLength {
		return 0, false
	}
	if s[5] == '2' && (s[6] == 'A' || s[6] == 'B') {
		s = s[:5] + map[byte]string{'A': "19", 'B': "18"}[s[6]] + s[7:]
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}
