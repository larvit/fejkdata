package jsonvalue

import "testing"

func TestDecodeNamesTheLineAndColumnOfASyntaxError(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{"a trailing comma", "{\"a\": 1,}", "line 1, column 9: invalid character '}' looking for beginning of object key string"},
		{"a missing comma on the third line", "{\n  \"a\": 1\n  \"b\": 2\n}", "line 3, column 3: invalid character '\"' after object key:value pair"},
		{"an empty input", "", "line 1, column 1: unexpected end of JSON input"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Decode([]byte(c.in)); err == nil || err.Error() != c.want {
				t.Errorf("Decode(%q) = %v, want %q", c.in, err, c.want)
			}
		})
	}
}
