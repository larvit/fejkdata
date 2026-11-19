package fejkdata

import (
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

const licenseLineWidth = 100

func TestDataLicensesLinesWrap(t *testing.T) {
	text, err := os.ReadFile("DATA-LICENSES.md")
	if err != nil {
		t.Fatal(err)
	}
	for n, line := range strings.Split(string(text), "\n") {
		cut := strings.LastIndex(line, " ")
		if cut >= 0 && utf8.RuneCountInString(line[:cut]) > licenseLineWidth {
			t.Errorf("DATA-LICENSES.md:%d runs past %d characters beside a space; wrap it", n+1, licenseLineWidth)
		}
	}
}
