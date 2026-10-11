package jsonvalue

import "testing"

func TestDecodeNamesTheLineAndColumnOfASyntaxError(t *testing.T) {
	for in, want := range map[string]string{
		"{\"a\": 1,}":                  "line 1, column 9: invalid character '}' looking for beginning of object key string",
		"{\n  \"a\": 1\n  \"b\": 2\n}": "line 3, column 3: invalid character '\"' after object key:value pair",
		"":                             "line 1, column 1: unexpected end of JSON input",
	} {
		if _, err := Decode([]byte(in)); err == nil || err.Error() != want {
			t.Errorf("Decode(%q) = %v, want %q", in, err, want)
		}
	}
}
