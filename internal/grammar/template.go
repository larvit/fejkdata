package grammar

import (
	"encoding/json"
	"fmt"
	"strings"
)

func IsTemplate(arg string) (bool, error) {
	if strings.ContainsRune(arg, '{') || (isJSONStart(strings.TrimSpace(arg)) && json.Valid([]byte(arg))) {
		if path, lone := loneReference(arg); lone {
			return false, pathAdvice(path, fmt.Sprintf("%s is the path %s written as a template", arg, path))
		}
		return true, nil
	}
	if i := strings.IndexAny(arg, `}"`); i >= 0 {
		return false, fmt.Errorf("%q holds a %q, which no path may, and it is not valid JSON, so it names no template either", arg, arg[i:i+1])
	}
	if strings.HasPrefix(strings.TrimSpace(arg), "[") {
		return false, fmt.Errorf(`%q starts with "[" and is not valid JSON, so it names no template; a path starts with a name`, arg)
	}
	if _, err := SplitPath(arg); err != nil {
		return false, err
	}
	if path := strings.TrimLeft(arg, "/"); path != arg && path != "" {
		return false, pathAdvice(path, fmt.Sprintf("path %s starts with /, and every path starts at the root already", arg))
	}
	return false, nil
}

func isJSONStart(arg string) bool {
	return strings.HasPrefix(arg, "[") || strings.HasPrefix(arg, `"`)
}

// pathAdvice refuses a spelling of path, naming path to write, or why no name can spell it.
func pathAdvice(path, refusal string) error {
	if err := CheckPathNames(path); err != nil {
		return err
	}
	return fmt.Errorf("%s; write %s", refusal, path)
}

// loneReference is the path a template spells when the value it holds — a format string, a
// JSON string, or an object holding only a format — is one reference token and nothing else.
func loneReference(arg string) (string, bool) {
	var raw any
	if json.Unmarshal([]byte(arg), &raw) != nil {
		raw = arg
	}
	if m, isObject := raw.(map[string]any); isObject && len(m) == 1 {
		raw = m["format"]
	}
	format, isString := raw.(string)
	if !isString {
		return "", false
	}
	toks, err := ParseFormat(format)
	if err != nil {
		return "", false
	}
	body, lone := loneRef(toks)
	if !lone {
		return "", false
	}
	if strings.HasPrefix(body, "/") {
		body = "/" + strings.TrimLeft(body, "/")
	}
	_, path, err := RefShape(body)
	return path, err == nil
}

// loneRef is the reference a format of one reference token and nothing else reads.
func loneRef(toks []Token) (string, bool) {
	if len(toks) != 1 || toks[0].Kind != NameRead || len(toks[0].Arms) != 1 || !IsRef(toks[0].Arms[0]) {
		return "", false
	}
	return toks[0].Arms[0], true
}
