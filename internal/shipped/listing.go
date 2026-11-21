package shipped

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

const (
	// PathsFile is the listing as JSON, which programs read.
	PathsFile = "data-listing/paths.json"
	// KindsFile is the hand-written kinds and source of each category, and of each field that
	// differs from its category.
	KindsFile = "data-listing/kinds-and-sources.tsv"
	// DocFile is the listing as Markdown.
	DocFile = "docs/shipped-paths.md"
	// ListingShapeFile is the pin of the listing's keys and kind words.
	ListingShapeFile = "testdata/listing_shape.txt"

	docExampleRunes = 48
)

// KindWords is every kind a path may have, in the order a path lists its kinds.
var KindWords = []string{"real", "valid", "invented"}

// Path is one line of the listing. Its JSON keys and KindWords are the listing's contract: the
// README's Versioning table says what a change to either is, and ListingShape pins both.
type Path struct {
	Path    string   `json:"path"`
	Kind    []string `json:"kind"`
	Example string   `json:"example"`
	Source  string   `json:"source"`
}

// ListingState is the draw state a listed path's example is drawn from: seed 1 on a fixed day.
func ListingState() *drawstate.State {
	return drawstate.New(1, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))
}

type kindRow struct {
	source string
	kinds  []string
	used   bool
}

// readKinds parses KindsFile. A row is tab-separated: a dot path, its kinds joined by ", " in
// the order of KindWords, and its source.
func readKinds(tsv []byte) (map[string]*kindRow, []string) {
	rows := map[string]*kindRow{}
	var mistakes []string
	for i, line := range strings.Split(strings.TrimSuffix(string(tsv), "\n"), "\n") {
		where := fmt.Sprintf("%s:%d", KindsFile, i+1)
		cells := strings.Split(line, "\t")
		if len(cells) != 3 || cells[0] == "" || cells[2] == "" {
			mistakes = append(mistakes, where+": want path, kinds and source, tab-separated, none empty")
			continue
		}
		if rows[cells[0]] != nil {
			mistakes = append(mistakes, where+": "+cells[0]+" is listed twice")
		}
		kinds := strings.Split(cells[1], ", ")
		if !inKindOrder(kinds) {
			mistakes = append(mistakes, where+": kinds are "+strings.Join(KindWords, ", ")+", each at most once and in that order")
		}
		rows[cells[0]] = &kindRow{kinds: kinds, source: cells[2]}
	}
	return rows, mistakes
}

func inKindOrder(kinds []string) bool {
	last := -1
	for _, k := range kinds {
		at := slices.Index(KindWords, k)
		if at <= last {
			return false
		}
		last = at
	}
	return true
}

// Listing is every path root lists, with the kinds and source kinds gives it, in KindsFile's
// format: the row of the path itself, else of the category its value is drawn from. Each
// example is drawn as a user draws it, from a state of its own that state returns, so adding a
// path moves no other path's example. The error names every path without a row, every row
// without a path and every mistake in kinds.
func Listing(root *tree.Folder, kinds []byte, state func() *drawstate.State) ([]Path, error) {
	rows, mistakes := readKinds(kinds)
	var out []Path
	for _, path := range engine.List(root) {
		listed, err := listPath(root, rows, path, state())
		if err != nil {
			mistakes = append(mistakes, err.Error())
			continue
		}
		out = append(out, listed)
	}
	for p, r := range rows {
		if !r.used {
			mistakes = append(mistakes, KindsFile+": "+p+" is no shipped path")
		}
	}
	if len(mistakes) > 0 {
		sort.Strings(mistakes)
		return nil, errors.New(strings.Join(mistakes, "\n"))
	}
	return out, nil
}

func listPath(root *tree.Folder, rows map[string]*kindRow, path string, s *drawstate.State) (Path, error) {
	row := rows[path]
	if row == nil {
		category, err := engine.DrawnFrom(root, path)
		if err != nil {
			return Path{}, err
		}
		row = rows[category]
	}
	if row == nil {
		return Path{}, errors.New("no row in " + KindsFile + " for " + path)
	}
	row.used = true
	template, err := engine.LoadTemplate(root, "{/"+path+"}", "template")
	if err != nil {
		return Path{}, fmt.Errorf("%s: %w", path, err)
	}
	return Path{Path: path, Kind: row.kinds, Example: template.Draw(s), Source: row.source}, nil
}

// ListingFiles is the listing as PathsFile and as DocFile, which holds one folded table per
// module.
func ListingFiles(listed []Path, namespaces []string) []File {
	return []File{{PathsFile, listingJSON(listed)}, {DocFile, listingMarkdown(listed, namespaces)}}
}

func listingJSON(listed []Path) []byte {
	var b bytes.Buffer
	b.WriteString("[")
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	for i, l := range listed {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("\n\t")
		_ = enc.Encode(l)
		b.Truncate(b.Len() - 1)
	}
	b.WriteString("\n]\n")
	return b.Bytes()
}

func listingMarkdown(listed []Path, namespaces []string) []byte {
	namespaces = slices.Clone(namespaces)
	sort.Strings(namespaces)
	var b strings.Builder
	b.WriteString("# Every shipped path\n\nThe README's [Every shipped path](../README.md#every-shipped-path) says what a kind, an example and a source are.\n\n")
	for i, n := range namespaces {
		held := slices.DeleteFunc(slices.Clone(listed), func(l Path) bool { return holder(namespaces, l.Path) != i })
		if len(held) == 0 {
			continue
		}
		fmt.Fprintf(&b, "<details>\n<summary><code>%s</code>, %d paths</summary>\n\n", n, len(held))
		b.WriteString("| Path | Kind | Example | Source |\n|---|---|---|---|\n")
		for _, l := range held {
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", l.Path, strings.Join(l.Kind, ", "), exampleCell(l.Example), strings.ReplaceAll(l.Source, "|", `\|`))
		}
		b.WriteString("\n</details>\n\n")
	}
	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}

// exampleCell is an example as a table cell: cut to docExampleRunes, on one line, in a code span.
func exampleCell(example string) string {
	if utf8.RuneCountInString(example) > docExampleRunes {
		example = string([]rune(example)[:docExampleRunes]) + "…"
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

// ListingShape is the pin of the listing's contract: a line for each key of a listed path, then
// one for each kind word.
func ListingShape() File {
	var b strings.Builder
	for i := range reflect.TypeFor[Path]().NumField() {
		b.WriteString("key " + reflect.TypeFor[Path]().Field(i).Tag.Get("json") + "\n")
	}
	for _, k := range KindWords {
		b.WriteString("kind " + k + "\n")
	}
	return File{ListingShapeFile, []byte(b.String())}
}
