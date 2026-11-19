package fejkdata

import (
	"fmt"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// Per python-stdnum's lt.asmens, 33309240064 is a valid code: the first ten digits weighted 1 to 9
// and 1 add up to 109, which is 10 modulo 11, so the second weights give its check digit, 4. The
// same ten digits with another check digit are not valid.
func TestValidAsmensKodasAppliesTheTwoWeights(t *testing.T) {
	for code, want := range map[string]bool{
		"33309240064": true, "33309240063": false, "33309240060": false, "3330924006": false, "333092400640": false,
		"93309240064": false, "03309240064": false, "73309240064": false, "33313240064": false, "33309320064": false,
	} {
		if got := fejktest.ValidAsmensKodas(code); got != want {
			t.Errorf("ValidAsmensKodas(%q) = %v, want %v", code, got, want)
		}
	}
}

const asmensKodasTemplate = `{"format":"{asmenskodas(sex,d)}","d":"%s","sex":"%s"}`

func asmensKodasOf(t *testing.T, f *Generator, date, sex string) string {
	t.Helper()
	return mustRender(t, f, fmt.Sprintf(asmensKodasTemplate, date, sex))
}

// The first digit is 1 for a man and 2 for a woman born in the 1800s, 3 and 4 in the 1900s and 5
// and 6 in the 2000s, as the Lithuanian Wikipedia lists them; the date
// follows as year, month and day of two digits each.
func TestBuiltinAsmenskodasStartsWithTheDigitOfItsSexAndCentury(t *testing.T) {
	f := emptyGenerator(21)
	for _, c := range []struct{ date, sex, want string }{
		{"18990101", "1", "1990101"}, {"18990101", "2", "2990101"},
		{"19330924", "1", "3330924"}, {"19330924", "2", "4330924"},
		{"20240229", "1", "5240229"}, {"20240229", "2", "6240229"},
		{"20991231", "2", "6991231"},
	} {
		t.Run(c.date+" "+c.sex, func(t *testing.T) {
			got := asmensKodasOf(t, f, c.date, c.sex)
			if len(got) != 11 || got[:7] != c.want || !fejktest.ValidAsmensKodas(got) {
				t.Errorf("asmenskodas for %s and %s = %q, want a valid code starting with %s", c.date, c.sex, got, c.want)
			}
		})
	}
}
