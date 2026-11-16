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
