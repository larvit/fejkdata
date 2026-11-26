package jsonvalue

import (
	"encoding/json"
	"strings"
	"testing"
)

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

func TestDecodeKeepsAnObjectsKeysInTheOrderWritten(t *testing.T) {
	v, err := Decode([]byte(`{"b": 1, "a": {"z": [{"y": null, "x": true}], "c": "s"}, "b": 2}`))
	if err != nil {
		t.Fatal(err)
	}
	top := v.Any().(Object)
	if got := strings.Join(top.Keys, " "); got != "b a" {
		t.Errorf("top keys = %q, want \"b a\": a key written twice keeps its first place", got)
	}
	if got := top.Fields["b"]; got != json.Number("2") {
		t.Errorf("b = %v, want 2, the value written last", got)
	}
	a := top.Fields["a"].(Object)
	if got := strings.Join(a.Keys, " "); got != "z c" {
		t.Errorf("a's keys = %q, want \"z c\"", got)
	}
	inList := a.Fields["z"].([]any)[0].(Object)
	if got := strings.Join(inList.Keys, " "); got != "y x" {
		t.Errorf("keys of the object in a list = %q, want \"y x\"", got)
	}
}

func TestSortedObjectListsItsKeysInNameOrder(t *testing.T) {
	o := SortedObject(map[string]Value{"b": String("1"), "a": String("2")}).Any().(Object)
	if got := strings.Join(o.Keys, " "); got != "a b" || o.Fields["b"] != "1" {
		t.Errorf("SortedObject = %v, want keys a b and b holding 1", o)
	}
}
