package grammar

import "testing"

func TestRefShapeSplitsTheSigil(t *testing.T) {
	for ref, want := range map[string][2]string{
		"/a.b":  {"/", "a.b"},
		".a":    {".", "a"},
		"..a.b": {"..", "a.b"},
	} {
		sigil, rest, err := RefShape(ref)
		if err != nil || sigil != want[0] || rest != want[1] {
			t.Errorf("RefShape(%q) = %q, %q, %v, want %q", ref, sigil, rest, err, want)
		}
	}
	for _, ref := range []string{"/", "//a", ".../a", "/a..b.", "/a.[b]"} {
		if _, _, err := RefShape(ref); err == nil {
			t.Errorf("RefShape(%q) = nil error, want one", ref)
		}
	}
}
