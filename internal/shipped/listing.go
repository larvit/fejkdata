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
	"github.com/larvit/fejkdata/internal/invariant"
)

const (
	// PathsFile is the listing as JSON, which programs read.
	PathsFile = "data-listing/paths.json"
	// KindsFile is the hand-written kinds and source of each category, and of each field that
	// differs from its category.
	KindsFile = "data-listing/kinds-and-sources.tsv"
	// DocFile is the listing as Markdown.
	DocFile = "docs/shipped-paths.md"
	// TypesFile is the listing as a TypeScript union of the paths, which the npm package ships.
	TypesFile = "npm-package/shipped-paths.d.ts"
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

// Listing is every path the tree lists, with its kinds and source from the rows of kinds, a file in
// KindsFile's format: the path's own row, else the row of the category its value is drawn from.
// Each example is drawn as a user draws it, from a fresh state that each call of state returns, so
// adding a path moves no other path's example. The error names every path without a row, every row
// without a path and every mistake in kinds.
func (d *Data) Listing(kinds []byte, state func() *drawstate.State) ([]Path, error) {
	rows, mistakes := readKinds(kinds)
	used := map[string]bool{}
	var out []Path
	for _, path := range engine.List(&d.root) {
		row, err := rowOf(&d.root, rows, path)
		if err != nil {
			mistakes = append(mistakes, err.Error())
			continue
		}
		used[row] = true
		template, err := engine.LoadTemplate(&d.root, "{/"+path+"}", "template")
		if err != nil {
			panic(invariant.Broken("the listed path %s does not load as a reference: %v", path, err))
		}
		out = append(out, Path{Path: path, Kind: rows[row].kinds, Example: template.Draw(state()), Source: rows[row].source})
	}
	for p := range rows {
		if !used[p] {
			mistakes = append(mistakes, KindsFile+": "+p+" is no shipped path")
		}
	}
	if len(mistakes) > 0 {
		sort.Strings(mistakes)
		return nil, errors.New(strings.Join(mistakes, "\n"))
	}
	return out, nil
}

// rowOf is the path of the row in rows that lists path. It is the path's own row, else the row
// of the field the path ends in, at the category its value is drawn from, so a field reached
// through a linked table needs no second row, else that category's row.
func rowOf(root *tree.Folder, rows map[string]*kindRow, path string) (string, error) {
	if rows[path] != nil {
		return path, nil
	}
	category, err := engine.DrawnFrom(root, path)
	if err != nil {
		return "", err
	}
	if field := category + path[strings.LastIndex(path, "."):]; rows[field] != nil && endsAtField(root, path, category) {
		return field, nil
	}
	if rows[category] == nil {
		return "", errors.New("no row in " + KindsFile + " for " + path)
	}
	return category, nil
}

// endsAtField reports whether path's last step is a field of category, where its value is drawn,
// and not a table linked to it: the path before that step is then drawn from category too.
func endsAtField(root *tree.Folder, path, category string) bool {
	before, err := engine.DrawnFrom(root, path[:strings.LastIndex(path, ".")])
	return err == nil && before == category
}

// ListingFiles is the listing as PathsFile, as DocFile, which holds one folded table per
// module, and as TypesFile.
func (d *Data) ListingFiles(listed []Path) []File {
	return []File{{PathsFile, listingJSON(listed)}, {DocFile, d.listingMarkdown(listed)}, {TypesFile, listingTypes(listed)}}
}

func listingTypes(listed []Path) []byte {
	var b strings.Builder
	b.WriteString("// Code generated by go generate; DO NOT EDIT.\n\n")
	b.WriteString("/** A path of the shipped data. Its comment shows a value it prints. */\n")
	b.WriteString("export type ShippedPath =")
	for _, l := range listed {
		fmt.Fprintf(&b, "\n\t%s\n\t| %q", typeComment(l.Example), l.Path)
	}
	b.WriteString(";\n")
	return []byte(b.String())
}

// typeComment is an example as a doc comment: whole, on one line, in a code span. A `*/` in it
// is written `*\/`, since the code span has no escape that would keep the comment open.
func typeComment(example string) string {
	return "/** " + codeSpan(strings.NewReplacer("\r", "", "\n", " ⏎ ", "*/", `*\/`).Replace(example)) + " */"
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

func (d *Data) listingMarkdown(listed []Path) []byte {
	namespaces := slices.Clone(d.namespaces)
	sort.Strings(namespaces)
	var b strings.Builder
	b.WriteString("# Every shipped path\n\nThe README's [Every shipped path](../README.md#every-shipped-path) says what a kind, an example and a source are.\n\n")
	for _, n := range namespaces {
		held := slices.DeleteFunc(slices.Clone(listed), func(l Path) bool { return d.namespaces[d.holder(l.Path)] != n })
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
	return codeSpan(strings.NewReplacer("\r", "", "\n", " ⏎ ", "|", `\|`).Replace(example))
}

func codeSpan(text string) string {
	if text == "" {
		return "(empty)"
	}
	fence := "`"
	if strings.Contains(text, "`") {
		fence = "``"
		text = " " + text + " "
	}
	return fence + text + fence
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
