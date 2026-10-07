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
	path, err := CallerPath(arg)
	if err != nil {
		return false, err
	}
	_, err = SplitPath(path)
	return false, err
}

// CallerPath is a path as a caller passed it, without the one / it may start with, as a
// reference does.
func CallerPath(path string) (string, error) {
	rest := strings.TrimPrefix(path, "/")
	if strings.HasPrefix(rest, "/") {
		return "", fmt.Errorf("path %s starts with more than one /; write /%s", path, strings.TrimLeft(rest, "/"))
	}
	return rest, nil
}

func isJSONStart(arg string) bool {
	return strings.HasPrefix(arg, "[") || strings.HasPrefix(arg, `"`)
}
