package data_test

import (
	"strconv"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// nameCounts reads a locale's names table as the count of each name, per sex for first names and
// under "" for last names.
func nameCounts(t *testing.T, locale, file string) map[string]map[string]int {
	t.Helper()
	out := map[string]map[string]int{}
	for _, row := range fejktest.TSVRows(t, locale+"/"+file+".tsv") {
		n, err := strconv.Atoi(row["count"])
		if err != nil || n < 1 {
			t.Fatalf("%s row %v has no count", file, row)
		}
		sex := row["sex"]
		if out[sex] == nil {
			out[sex] = map[string]int{}
		}
		if _, dup := out[sex][row["name"]]; dup {
			t.Fatalf("%s holds %q twice for sex %q", file, row["name"], sex)
		}
		out[sex][row["name"]] = n
	}
	return out
}

func nameColumnSet(t *testing.T, file string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, row := range fejktest.TSVRows(t, file) {
		out[row["name"]] = true
	}
	return out
}

// lastNameCounts reads a locale's last names as the count of each name, whether or not the table has a sex column.
func lastNameCounts(t *testing.T, locale string) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, names := range nameCounts(t, locale, "last-name") {
		for name, n := range names {
			out[name] = max(out[name], n)
		}
	}
	return out
}

func TestNameTablesOfCountedBearersHoldNoNameUnderFiveBearers(t *testing.T) {
	for _, locale := range []string{"en_US", "fi_FI", "fr_FR", "sv_SE"} {
		for _, file := range []string{"first-name", "last-name"} {
			t.Run(locale+"/"+file, func(t *testing.T) {
				for sex, names := range nameCounts(t, locale, file) {
					for name, n := range names {
						if n < 5 {
							t.Errorf("%s %q of sex %q is borne by %d, want at least 5", file, name, sex, n)
						}
					}
				}
			})
		}
	}
}
