package grammar

import (
	"fmt"
	"strings"
)

// A reference names a node by path rather than as a sibling field: {/a.b} from the
// data root, {.a} from the folder this file sits in, {..a} from the folder above.
func IsRef(name string) bool { return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "/") }

// RefShape splits a reference into its sigil and the dotted path after it.
func RefShape(name string) (sigil, rest string, err error) {
	switch {
	case strings.HasPrefix(name, "/"):
		sigil, rest = "/", name[1:]
	case strings.HasPrefix(name, ".."):
		sigil, rest = "..", name[2:]
	default:
		sigil, rest = ".", name[1:]
	}
	if rest == "" {
		return "", "", fmt.Errorf("reference has no path")
	}
	if strings.HasPrefix(rest, "/") {
		return "", "", fmt.Errorf("the path after %s starts at a name, not a /; write {%s%s}", sigil, sigil, rest[1:])
	}
	if strings.HasPrefix(rest, ".") {
		return "", "", fmt.Errorf("a reference starts with / (the root), . (this folder) or .. (the folder above)")
	}
	segs, err := SplitPath(rest)
	if err != nil {
		return "", "", err
	}
	if err := CheckSegments(segs); err != nil {
		return "", "", err
	}
	return sigil, rest, nil
}

// IsLonePath reports whether body, the text after a reference's /, is one path and nothing
// else: written as {/body}, it parses to one path read of one reference.
func IsLonePath(body string) bool {
	toks, err := ParseFormat("{/" + body + "}")
	if err != nil || len(toks) != 1 || toks[0].Kind != PathRead || len(toks[0].Arms) != 1 {
		return false
	}
	sigil, path, err := RefShape(toks[0].Arms[0])
	return err == nil && sigil == "/" && path == body
}
