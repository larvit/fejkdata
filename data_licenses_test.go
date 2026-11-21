package fejkdata

import (
	"encoding/json"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/shipped"
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
			t.Errorf("DATA-LICENSES.md:%d has a space after column %d; wrap the line there", n+1, licenseLineWidth)
		}
	}
}

var (
	licenseFiles = regexp.MustCompile("(?ms)^- Files: (.*?)^- ")
	sourceWord   = regexp.MustCompile(`[^\s,;:()]+`)
	properWord   = regexp.MustCompile(`^\p{Lu}|\d`)
	possessive   = regexp.MustCompile(`['’]s?$`)
)

// licenseEntries is the text of each DATA-LICENSES.md entry, by every rows file it covers. An entry
// names its files in its heading, or under Files, and a name without a folder takes the folder of
// the name before it.
func licenseEntries(t *testing.T) map[string]string {
	t.Helper()
	text, err := os.ReadFile("DATA-LICENSES.md")
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]string{}
	for _, entry := range strings.Split(string(text), "\n## ")[1:] {
		names, _, _ := strings.Cut(entry, "\n")
		if m := licenseFiles.FindStringSubmatch(entry); m != nil {
			names = m[1]
		}
		folder := ""
		for _, m := range codeSpan.FindAllStringSubmatch(names, -1) {
			name := m[1]
			if strings.Contains(name, "/") {
				folder = path.Dir(name)
			} else {
				name = folder + "/" + name
			}
			entries[name] += entry
		}
	}
	return entries
}

// rowsFileOf is the rows file the category at a dot path reads, from data/, or "" where the path
// is no category with one.
func rowsFileOf(category string) string {
	file := strings.ReplaceAll(category, ".", "/")
	b, err := os.ReadFile("data/" + file + ".json")
	if err != nil {
		return ""
	}
	var c struct{ Rows string }
	if json.Unmarshal(b, &c) != nil || c.Rows == "" {
		return ""
	}
	return path.Dir(file) + "/" + c.Rows
}

func TestAListedSourceNamesWhatItsRowsFilesLicenceEntryNames(t *testing.T) {
	kinds, err := os.ReadFile(shipped.KindsFile)
	if err != nil {
		t.Fatal(err)
	}
	entries := licenseEntries(t)
	checked := 0
	for n, line := range strings.Split(strings.TrimSuffix(string(kinds), "\n"), "\n") {
		cells := strings.Split(line, "\t")
		entry := entries[rowsFileOf(cells[0])]
		if entry == "" {
			continue
		}
		checked++
		for _, word := range sourceWord.FindAllString(cells[2], -1) {
			word = possessive.ReplaceAllString(word, "")
			if properWord.MatchString(word) && !strings.Contains(entry, word) {
				t.Errorf("%s:%d: the source of %s names %q, which the DATA-LICENSES.md entry of its rows file does not; spell it as the register itself does, in both", shipped.KindsFile, n+1, cells[0], word)
			}
		}
	}
	if checked < 100 {
		t.Fatalf("only %d rows of %s read a rows file with a DATA-LICENSES.md entry; the check lost its rows", checked, shipped.KindsFile)
	}
}
