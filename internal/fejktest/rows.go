package fejktest

import (
	"os"
	"strings"
	"testing"
)

func TSVRows(t *testing.T, path string) []map[string]string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(src)), "\n")
	head := strings.Split(lines[0], "\t")
	rows := make([]map[string]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		cells := strings.Split(line, "\t")
		row := make(map[string]string, len(head))
		for i, name := range head {
			row[name] = cells[i]
		}
		rows = append(rows, row)
	}
	return rows
}
