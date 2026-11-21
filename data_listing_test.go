package fejkdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/fejktest"
)

const (
	listingPaths       = "data-listing/paths.json"
	listingKinds       = "data-listing/kinds-and-sources.tsv"
	readmeListingStart = "<!-- data-listing:start -->"
	readmeListingEnd   = "<!-- data-listing:end -->"
	readmeExampleRunes = 48
)

var listingKindWords = []string{"real", "valid", "invented"}

type listedPath struct {
	Path    string   `json:"path"`
	Kind    []string `json:"kind"`
	Example string   `json:"example"`
	Source  string   `json:"source"`
}

type kindRow struct {
	pattern string
	source  string
	kinds   []string
	used    bool
}

// readKindRows parses kinds-and-sources.tsv: a category's dot path, or `*.` and the last segments of
// several, then its kinds joined by ", ", then its source.
func readKindRows() ([]*kindRow, []string) {
	raw, err := os.ReadFile(listingKinds)
	if err != nil {
		return nil, []string{err.Error()}
	}
	var rows []*kindRow
	var errs []string
	seen := map[string]bool{}
	for i, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		where := fmt.Sprintf("%s:%d", listingKinds, i+1)
		cells := strings.Split(line, "\t")
		if len(cells) != 3 || cells[0] == "" || cells[2] == "" {
			errs = append(errs, where+": want category, kinds and source, tab-separated, none empty")
			continue
		}
		if seen[cells[0]] {
			errs = append(errs, where+": "+cells[0]+" is listed twice")
		}
		seen[cells[0]] = true
		kinds := strings.Split(cells[1], ", ")
		for _, k := range kinds {
			if !slices.Contains(listingKindWords, k) {
				errs = append(errs, where+": kind "+k+" is none of "+strings.Join(listingKindWords, ", "))
			}
		}
		rows = append(rows, &kindRow{pattern: cells[0], kinds: kinds, source: cells[2]})
	}
	return rows, errs
}

// rowFor is the row naming category exactly, else the wildcard row with the most segments that ends it.
func rowFor(rows []*kindRow, category string) *kindRow {
	var best *kindRow
	for _, r := range rows {
		switch {
		case r.pattern == category:
			return r
		case strings.HasPrefix(r.pattern, "*.") && strings.HasSuffix(category, r.pattern[1:]):
			if best == nil || len(r.pattern) > len(best.pattern) {
				best = r
			}
		}
	}
	return best
}

// shippedListing is every path the whole shipped tree lists, each drawn once from a state of its own, so
// adding a path moves no other path's example.
func shippedListing(t *testing.T) ([]listedPath, []string, []string) {
	t.Helper()
	f := newShippedWhole(t)
	rows, errs := readKindRows()
	categories := loadedCategories(f)
	var out []listedPath
	for _, path := range f.List() {
		category := ""
		for _, c := range categories {
			if (path == c || strings.HasPrefix(path, c+".")) && len(c) > len(category) {
				category = c
			}
		}
		row := rowFor(rows, category)
		if category == "" || row == nil {
			if msg := "no row in " + listingKinds + " for " + category; !slices.Contains(errs, msg) {
				errs = append(errs, msg)
			}
			continue
		}
		row.used = true
		example, err := engine.Fake(drawstate.New(1, fejktest.TestDay), &f.root, path)
		if err != nil {
			errs = append(errs, path+": "+err.Error())
			continue
		}
		out = append(out, listedPath{Path: path, Kind: row.kinds, Example: example, Source: row.source})
	}
	for _, r := range rows {
		if !r.used {
			errs = append(errs, listingKinds+": "+r.pattern+" names no shipped category")
		}
	}
	return out, errs, categories
}

func listingJSON(listed []listedPath) []byte {
	var b bytes.Buffer
	b.WriteString("[")
	for i, l := range listed {
		line, err := marshalUnescaped(l)
		if err != nil {
			panic(err)
		}
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("\n\t" + line)
	}
	b.WriteString("\n]\n")
	return b.Bytes()
}

func marshalUnescaped(v any) (string, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimSuffix(b.String(), "\n"), nil
}

// readmeListing is the README's table of every category, one folded table a module.
func readmeListing(t *testing.T, listed []listedPath, categories []string) string {
	t.Helper()
	byModule := map[string][]listedPath{}
	for _, l := range listed {
		for _, m := range shippedModules(t) {
			if m.holds(l.Path) && slices.Contains(categories, l.Path) {
				byModule[m.namespace] = append(byModule[m.namespace], l)
			}
		}
	}
	var namespaces []string
	for n := range byModule {
		namespaces = append(namespaces, n)
	}
	sort.Strings(namespaces)
	var b strings.Builder
	for _, n := range namespaces {
		fmt.Fprintf(&b, "<details>\n<summary><code>%s</code>, %d categories</summary>\n\n", n, len(byModule[n]))
		b.WriteString("| Path | Kind | Example | Source |\n|---|---|---|---|\n")
		for _, l := range byModule[n] {
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", l.Path, strings.Join(l.Kind, ", "), markdownCell(l.Example), tableCell(l.Source))
		}
		b.WriteString("\n</details>\n\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func tableCell(s string) string { return strings.ReplaceAll(s, "|", `\|`) }

func markdownCell(example string) string {
	if utf8.RuneCountInString(example) > readmeExampleRunes {
		example = string([]rune(example)[:readmeExampleRunes]) + "…"
	}
	example = strings.NewReplacer("\r", "", "\n", " ⏎ ", "|", `\|`).Replace(example)
	if example == "" {
		return "(empty)"
	}
	fence := "`"
	if strings.Contains(example, "`") {
		fence = "``"
		example = " " + example + " "
	}
	return fence + example + fence
}

func replaceBetween(doc, start, end, body string) (string, bool) {
	i, j := strings.Index(doc, start), strings.Index(doc, end)
	if i < 0 || j < i {
		return doc, false
	}
	return doc[:i+len(start)] + "\n\n" + body + "\n" + doc[j:], true
}

// REGENERATE=1 rewrites data-listing/paths.json and the README's table.
func TestDataListingIsCurrent(t *testing.T) {
	listed, errs, categories := shippedListing(t)
	if len(errs) > 0 {
		t.Fatal("the listing is not whole:\n" + strings.Join(errs, "\n"))
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	newReadme, ok := replaceBetween(string(readme), readmeListingStart, readmeListingEnd, readmeListing(t, listed, categories))
	if !ok {
		t.Fatalf("README.md needs %s and %s around the table of every category", readmeListingStart, readmeListingEnd)
	}
	paths := listingJSON(listed)
	if os.Getenv("REGENERATE") == "1" {
		if err := os.WriteFile(listingPaths, paths, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("README.md", []byte(newReadme), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	const regenerate = "regenerate it: docker compose run --rm --user \"$(id -u):$(id -g)\" generate"
	if want, err := os.ReadFile(listingPaths); err != nil || string(want) != string(paths) {
		t.Errorf("%s is stale; %s", listingPaths, regenerate)
	}
	if newReadme != string(readme) {
		t.Errorf("README.md's table of every category is stale; %s", regenerate)
	}
}
