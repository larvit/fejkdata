package fejkdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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
	listingPaths        = "data-listing/paths.json"
	listingKinds        = "data-listing/kinds-and-sources.tsv"
	listingDoc          = "docs/shipped-paths.md"
	listingExampleRunes = 48
	githubRenderLimit   = 500 * 1024
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

func lastSegment(path string) string { return path[strings.LastIndex(path, ".")+1:] }

// categoryOf is the category a path's value comes from: the deepest category it lies under, then each
// table linked below that category which the path names, since a linked table is a category of its own.
func categoryOf(lines []engine.IndexLine, path string) string {
	category := ""
	for _, l := range lines {
		if (path == l.Path || strings.HasPrefix(path, l.Path+".")) && len(l.Path) > len(category) {
			category = l.Path
		}
	}
	for category != "" && path != category {
		next := strings.SplitN(strings.TrimPrefix(path, category+"."), ".", 2)[0]
		want := strings.TrimSuffix(category, lastSegment(category)) + next
		i := slices.IndexFunc(lines, func(l engine.IndexLine) bool {
			return l.Path == want && l.Entry.Parent == lastSegment(category)
		})
		if i < 0 {
			break
		}
		category = lines[i].Path
	}
	return category
}

// rowFor is the row of the path itself, else of the category its value comes from.
func rowFor(rows map[string]*kindRow, lines []engine.IndexLine, path string) *kindRow {
	if r := rows[path]; r != nil {
		return r
	}
	return rows[categoryOf(lines, path)]
}

// shippedListing is every path the whole shipped tree lists, each drawn as a user draws it, from seed 1 on
// TestDay with a state of its own, so adding a path moves no other path's example.
func shippedListing(t *testing.T) ([]listedPath, []string) {
	t.Helper()
	f := newShippedWhole(t)
	rows, errs := readKindRows()
	lines := engine.Index(&f.root)
	var out []listedPath
	for _, path := range f.List() {
		row := rowFor(rows, lines, path)
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
	return out, errs
}

func TestCategoryOfStaysInTheCategorysFolder(t *testing.T) {
	lines := engine.Index(&newShippedWhole(t).root)
	for path, want := range map[string]string{
		"sv_SE.sex.personnummer":      "sv_SE.personnummer",
		"sv_SE.sex.first-name":        "sv_SE.first-name",
		"fi_FI.sex.first-name":        "fi_FI.first-name",
		"geo.SE.locality.postal-code": "geo.SE.postal-code",
		"geo.US.locality.street":      "geo.US.street",
		"sv_SE.person.first":          "sv_SE.person",
	} {
		if got := categoryOf(lines, path); got != want {
			t.Errorf("categoryOf(%s) = %s, want %s", path, got, want)
		}
	}
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

// listingMarkdown is docs/shipped-paths.md: every listed path, one folded table per module.
func listingMarkdown(t *testing.T, listed []listedPath) string {
	t.Helper()
	modules := shippedModules(t)
	byModule := map[string][]listedPath{}
	for _, l := range listed {
		for _, m := range modules {
			if m.holds(l.Path) {
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
	b.WriteString("# Every shipped path\n\nThe README's [Every shipped path](../README.md#every-shipped-path) says what a kind, an example and a source are.\n\n")
	for _, n := range namespaces {
		fmt.Fprintf(&b, "<details>\n<summary><code>%s</code>, %d paths</summary>\n\n", n, len(byModule[n]))
		b.WriteString("| Path | Kind | Example | Source |\n|---|---|---|---|\n")
		for _, l := range byModule[n] {
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", l.Path, strings.Join(l.Kind, ", "), markdownCell(l.Example), tableCell(l.Source))
		}
		b.WriteString("\n</details>\n\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func tableCell(s string) string { return strings.ReplaceAll(s, "|", `\|`) }

func markdownCell(example string) string {
	if utf8.RuneCountInString(example) > listingExampleRunes {
		example = string([]rune(example)[:listingExampleRunes]) + "…"
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

// REGENERATE=1 rewrites data-listing/paths.json and docs/shipped-paths.md.
func TestDataListingIsCurrent(t *testing.T) {
	listed, errs := shippedListing(t)
	if len(errs) > 0 {
		t.Fatal("the listing is not whole:\n" + strings.Join(errs, "\n"))
	}
	doc := listingMarkdown(t, listed)
	if len(doc) > githubRenderLimit {
		t.Errorf("%s is %d bytes, past the %d GitHub renders; split it per module", listingDoc, len(doc), githubRenderLimit)
	}
	generated := []struct{ file, body string }{
		{listingPaths, string(listingJSON(listed))},
		{listingDoc, doc},
	}
	regenerating := os.Getenv("REGENERATE") == "1"
	for _, g := range generated {
		file, body := g.file, g.body
		if regenerating {
			if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if have, err := os.ReadFile(file); err != nil || string(have) != body {
			t.Errorf("%s is stale; regenerate it: docker compose run --rm --user \"$(id -u):$(id -g)\" generate", file)
		}
	}
}

var codeSpan = regexp.MustCompile("`([^`]+)`")

// summaryNames is what the README's summary holds a line for: each folder of data/, and each country of data/geo.
func summaryNames(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("data")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() != "geo" {
			names = append(names, e.Name())
			continue
		}
		countries, err := os.ReadDir("data/geo")
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range countries {
			if c.IsDir() {
				names = append(names, "geo."+c.Name())
			}
		}
	}
	return names
}

func TestEachDataFolderHasASummaryLine(t *testing.T) {
	raw, err := os.ReadFile(listingPaths)
	if err != nil {
		t.Fatal(err)
	}
	var listed []listedPath
	if err := json.Unmarshal(raw, &listed); err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, l := range listed {
		known[l.Path] = true
	}
	section := readmeSection(t, readme(t), "### Every shipped path")
	for _, name := range summaryNames(t) {
		_, after, found := strings.Cut(section, "\n- `"+name+"`:")
		if !found {
			t.Errorf("the README's Every shipped path has no line starting - `%s`:", name)
			continue
		}
		line, _, _ := strings.Cut(after, "\n")
		spans := codeSpan.FindAllStringSubmatch(line, -1)
		if len(spans) == 0 {
			t.Errorf("the line for %s names no path in code", name)
		}
		for _, s := range spans {
			if !known[s[1]] {
				t.Errorf("the line for %s names %s, which --list does not print", name, s[1])
			}
		}
	}
}
