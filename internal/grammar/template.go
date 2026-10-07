package grammar

import (
	"encoding/json"
	"fmt"
	"strings"
)

func IsTemplate(arg string) (bool, error) {
	if strings.ContainsRune(arg, '{') || (isJSONStart(strings.TrimSpace(arg)) && json.Valid([]byte(arg))) {
		return true, nil
	}
	if i := strings.IndexAny(arg, `}"`); i >= 0 {
		return false, fmt.Errorf("%q holds a %q, which no path may, and it is not valid JSON, so it names no template either", arg, arg[i:i+1])
	}
	if strings.HasPrefix(strings.TrimSpace(arg), "[") {
		return false, fmt.Errorf(`%q starts with "[" and is not valid JSON, so it names no template; a path starts with a name`, arg)
	}
	_, err := SplitPath(CallerPath(arg))
	return false, err
}

// CallerPath is a path as a caller passed it, without the slashes it may start with.
func CallerPath(path string) string { return strings.TrimLeft(path, "/") }

func isJSONStart(arg string) bool {
	return strings.HasPrefix(arg, "[") || strings.HasPrefix(arg, `"`)
}
