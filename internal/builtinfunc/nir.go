package builtinfunc

import (
	"fmt"
	"strconv"
)

// nirLength is the number of characters of a NIR before its key: the sex, the year and month of
// birth, the place of birth and the order number.
const nirLength = 13

// nirKey is the key of the NIR made of the last nirLength letters and digits of emitted, as INSEE
// defines it: 97 less the remainder of the number divided by 97, where Corsica's 2A and 2B count as
// 19 and 18. It is empty where those characters are not that many, or not a number.
func nirKey(emitted string) string {
	n, ok := nirNumber(alnumTail(emitted, nirLength))
	if !ok {
		return ""
	}
	return fmt.Sprintf("%02d", 97-n%97)
}

// alnumTail is the last n ASCII letters and digits of s, fewer where s holds fewer.
func alnumTail(s string, n int) string {
	var kept []byte
	for i := range len(s) {
		if isAlnum(s[i]) {
			kept = append(kept, s[i])
		}
	}
	return string(kept[max(0, len(kept)-n):])
}

func isAlnum(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
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
