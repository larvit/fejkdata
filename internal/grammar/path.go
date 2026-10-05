package grammar

import (
	"fmt"
	"slices"
	"strings"
)

// SplitPath splits a dotted path into segments, a selector — [key or name] after a
// table's name — becoming a segment of its own, brackets kept, and so does each "..". A
// dot inside a selector is part of the key or name.
func SplitPath(path string) ([]string, error) {
	segs := make([]string, 0, strings.Count(path, ".")+2*strings.Count(path, "[")+1)
	start, open := 0, -1
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '[':
			if err := checkOpen(path, i, start, open); err != nil {
				return nil, err
			}
			segs = append(segs, path[start:i])
			start, open = i, i
		case ']':
			if err := checkClose(path, i, open); err != nil {
				return nil, err
			}
			segs = append(segs, path[start:i+1])
			open = -1
			segs, i = stepUpAt(segs, path, i+1)
			start = i + 1
		case '.':
			if open < 0 {
				segs = append(segs, path[start:i])
				segs, i = stepUpAt(segs, path, i)
				start = i + 1
			}
		}
	}
	if open >= 0 {
		return nil, fmt.Errorf(`%q opens a selector with "[" and never closes it with "]"`, path)
	}
	if start < len(path) || start == len(path) && (len(segs) == 0 || segs[len(segs)-1] != "..") {
		segs = append(segs, path[start:])
	}
	return segs, nil
}

// stepUpAt appends a ".." segment where the dot at i is the first of two, returning the
// index of the last dot it consumed.
func stepUpAt(segs []string, path string, i int) ([]string, int) {
	if i+1 < len(path) && path[i+1] == '.' {
		return append(segs, ".."), i + 1
	}
	return segs, i
}

// checkOpen refuses a "[" at i that opens no selector: one inside a selector, or one
// following no name.
func checkOpen(path string, i, start, open int) error {
	switch {
	case open >= 0:
		return fmt.Errorf(`%q holds a "[" inside a selector, which no key or name may`, path)
	case i == 0:
		return fmt.Errorf(`%q starts with "["; a path starts with a name, and a selector follows a table's name`, path)
	case i == start:
		return fmt.Errorf(`%q: a selector follows its table's name; write %s`, path, path[:i-1]+path[i:])
	}
	return nil
}

// checkClose refuses a "]" at i that closes no selector, closes an empty one, or is
// followed by anything but a dot or the end.
func checkClose(path string, i, open int) error {
	switch {
	case open < 0:
		return fmt.Errorf(`%q holds a "]" that closes no "["`, path)
	case i == open+1:
		return fmt.Errorf("%q holds an empty selector; a selector names a row by key or name", path)
	case i+1 < len(path) && path[i+1] != '.':
		return fmt.Errorf(`%q: a "]" ends its selector, so a dot or the end must follow it`, path)
	}
	return nil
}

// indexOutside is the first c in s outside a [selector], or -1.
func indexOutside(s string, c byte) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '[':
			depth++
		case s[i] == ']' && depth > 0:
			depth--
		case s[i] == c && depth == 0:
			return i
		}
	}
	return -1
}

// cutOutside cuts s around the first sep outside a [selector].
func cutOutside(s, sep string) (before, after string, found bool) {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '[':
			depth++
		case s[i] == ']' && depth > 0:
			depth--
		case depth == 0 && strings.HasPrefix(s[i:], sep):
			return s[:i], s[i+len(sep):], true
		}
	}
	return s, "", false
}

func splitOutside(s string, c byte) []string {
	var parts []string
	for {
		i := indexOutside(s, c)
		if i < 0 {
			return append(parts, s)
		}
		parts, s = append(parts, s[:i]), s[i+1:]
	}
}

func IsSelector(seg string) bool { return strings.HasPrefix(seg, "[") }

func HasSelector(segs []string) bool {
	for _, s := range segs {
		if IsSelector(s) {
			return true
		}
	}
	return false
}

func SelectorOf(seg string) string { return seg[1 : len(seg)-1] }

// NameSegments is the segments of a path that are names, its selectors and ".." left out.
func NameSegments(segs []string) []string {
	out := segs[:0:0]
	for _, s := range segs {
		if !IsSelector(s) && s != ".." {
			out = append(out, s)
		}
	}
	return out
}

// JoinSegments spells segments as a path: a selector attaches to the name before it, and
// so do a ".." and the name after it.
func JoinSegments(segs []string) string {
	var b strings.Builder
	for i, s := range segs {
		if b.Len() > 0 && !IsSelector(s) && s != ".." && segs[i-1] != ".." {
			b.WriteByte('.')
		}
		b.WriteString(s)
	}
	return b.String()
}

// CheckSegments refuses an unfinished path: "a." and "a...b" each have a segment naming
// nothing.
func CheckSegments(segs []string) error {
	if slices.Contains(segs, "") {
		return fmt.Errorf("path has an empty segment")
	}
	return nil
}
