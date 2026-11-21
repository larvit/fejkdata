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
	source string
	kinds  []string
	used   bool
}

// readKindRows parses kinds-and-sources.tsv. A row is tab-separated: a dot path, its kinds joined by ", "
// in the order of listingKindWords, and its source. A category has a row; a field of it has one where its
// kinds or source differ from the category's.
func readKindRows() (map[string]*kindRow, []string) {
	raw, err := os.ReadFile(listingKinds)
	if err != nil {
		return nil, []string{err.Error()}
	}
	rows := map[string]*kindRow{}
	var errs []string
	for i, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		where := fmt.Sprintf("%s:%d", listingKinds, i+1)
		cells := strings.Split(line, "\t")
		if len(cells) != 3 || cells[0] == "" || cells[2] == "" {
			errs = append(errs, where+": want path, kinds and source, tab-separated, none empty")
			continue
		}
		if rows[cells[0]] != nil {
			errs = append(errs, where+": "+cells[0]+" is listed twice")
		}
		kinds := strings.Split(cells[1], ", ")
		last := -1
		for _, k := range kinds {
			at := slices.Index(listingKindWords, k)
			if at <= last {
				errs = append(errs, where+": kinds are "+strings.Join(listingKindWords, ", ")+", each at most once and in that order")
				break
			}
			last = at
		}
		rows[cells[0]] = &kindRow{kinds: kinds, source: cells[2]}
	}
	return rows, errs
}

// rowFor is the row of the path itself, else of its nearest ancestor with one.
func rowFor(rows map[string]*kindRow, path string) *kindRow {
	for p := path; ; p = p[:strings.LastIndex(p, ".")] {
		if r := rows[p]; r != nil {
			return r
		}
		if !strings.Contains(p, ".") {
			return nil
		}
	}
}

// shippedListing is every path the whole shipped tree lists, each drawn as a user draws it, from seed 1 on
// TestDay with a state of its own, so adding a path moves no other path's example.
func shippedListing(t *testing.T) ([]listedPath, []string, []string) {
	t.Helper()
	f := newShippedWhole(t)
	rows, errs := readKindRows()
	var categories []string
	for _, l := range engine.Index(&f.root) {
		categories = append(categories, l.Path)
	}
	var out []listedPath
	for _, path := range f.List() {
		row := rowFor(rows, path)
		if row == nil {
			errs = append(errs, "no row in "+listingKinds+" for "+path)
			continue
		}
		row.used = true
		f.drawState = drawstate.New(1, fejktest.TestDay)
		example, err := f.FakeTemplate("{/" + path + "}")
		if err != nil {
			errs = append(errs, path+": "+err.Error())
			continue
		}
		out = append(out, listedPath{Path: path, Kind: row.kinds, Example: example, Source: row.source})
	}
	for p, r := range rows {
		if !r.used {
			errs = append(errs, listingKinds+": "+p+" is no shipped path")
		}
	}
	sort.Strings(errs)
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

// readmeListing is the README's table of every category, one folded table per module.
func readmeListing(t *testing.T, listed []listedPath, categories []string) string {
	t.Helper()
	modules := shippedModules(t)
	byModule := map[string][]listedPath{}
	for _, l := range listed {
		for _, m := range modules {
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
