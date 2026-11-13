package fejkdata

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

var (
	printedCall = regexp.MustCompile(`(?m)^.*Fake\w*\(.*//\s*".*$`)
	printedFake = regexp.MustCompile(`\.Fake\("([^"]+)"\)\s*//\s*("(?:[^"\\]|\\.)*")`)
)

func replayPrintedFakes(t *testing.T, name string, f *Generator, lines []string) {
	t.Helper()
	for _, line := range lines {
		c := printedFake.FindStringSubmatch(line)
		if c == nil {
			t.Fatalf("%s prints a value this test cannot replay; spell it f.Fake(\"path\") // \"value\": %s", name, line)
		}
		want, err := strconv.Unquote(c[2])
		if err != nil {
			t.Fatalf("%s prints %s: %v", name, c[2], err)
		}
		if got := fake(t, f, c[1]); got != want {
			t.Errorf("%s prints Fake(%q) = %q for seed 42, it renders %q", name, c[1], want, got)
		}
	}
}

// TestGoExamplesPrintWhatTheyRender replays each Go example's printed Fake calls on seed 42.
func TestGoExamplesPrintWhatTheyRender(t *testing.T) {
	for _, name := range []string{"doc.go", "README.md"} {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			lines := printedCall.FindAllString(string(src), -1)
			if lines == nil {
				t.Errorf("%s prints no seeded Fake, so nothing holds its example to a render", name)
			}
			f, err := New(withShipped(), WithSeed(42), onTestDay)
			if err != nil {
				t.Fatal(err)
			}
			replayPrintedFakes(t, name, f, lines)
		})
	}
}
