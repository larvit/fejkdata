package datafiles

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

// ManifestFile is the file at a source's root that describes the source.
const ManifestFile = ".fejkdata.json"

// Manifest is what a source's manifest says of it.
type Manifest struct {
	// Index is every category of the source by dot path; nil where the manifest has none.
	Index map[string]IndexEntry `json:"index"`
	// Reads is every module the source reads by default, directly or through another; the
	// error for a reference naming nothing names them, and nothing loads them.
	Reads []string `json:"reads"`
}

// IndexEntry is what an index says of a category: its parent table, "" for none, and the
// paths List advertises below it.
type IndexEntry struct {
	Parent string   `json:"parent,omitempty"`
	Paths  []string `json:"paths"`
}

// Manifest reads the manifest at the source's root, the zero Manifest where there is none.
func (s Source) Manifest() (Manifest, error) {
	if err := s.check(); err != nil {
		return Manifest{}, err
	}
	b, err := fs.ReadFile(s.fsys, ManifestFile)
	if errors.Is(err, fs.ErrNotExist) {
		return Manifest{}, nil
	}
	if err != nil {
		return Manifest{}, fmt.Errorf("%s: %w", s.ManifestPath(), err)
	}
	m, err := decodeManifest(b)
	if err != nil {
		return Manifest{}, fmt.Errorf("%s: %w", s.ManifestPath(), err)
	}
	return m, nil
}

// ManifestPath is the manifest's path as errors name it.
func (s Source) ManifestPath() string { return s.labelled(ManifestFile) }

// decodeManifest decodes a manifest, refusing any key or shape it does not define, and
// reports every mistake: first each entry's own, in path order, then each parent naming
// no entry.
func decodeManifest(b []byte) (Manifest, error) {
	raw, err := jsonvalue.Decode(b)
	if err != nil {
		return Manifest{}, err
	}
	top, isObject := raw.Any().(map[string]any)
	if !isObject {
		return Manifest{}, fmt.Errorf("a manifest must be an object, not %s", jsonvalue.Kind(raw.Any()))
	}
	var m Manifest
	var errs []error
	for _, k := range sortedKeys(top) {
		switch k {
		case "index":
			var indexErrs []error
			m.Index, indexErrs = decodeIndex(top[k])
			errs = append(errs, indexErrs...)
		case "reads":
			var err error
			m.Reads, err = decodeReads(top[k])
			errs = append(errs, err)
		default:
			errs = append(errs, fmt.Errorf("unknown key %q; a manifest holds index and reads", k))
		}
	}
	return m, errors.Join(errs...)
}

// decodeReads decodes the modules a source reads, each named by a non-empty string.
func decodeReads(v any) ([]string, error) {
	items, isList := v.([]any)
	if !isList {
		return nil, fmt.Errorf("reads must be a list, not %s", jsonvalue.Kind(v))
	}
	out := make([]string, 0, len(items))
	var errs []error
	for i, item := range items {
		module, isString := item.(string)
		switch {
		case !isString:
			errs = append(errs, fmt.Errorf("reads item %d must be a string, not %s", i+1, jsonvalue.Kind(item)))
		case module == "":
			errs = append(errs, fmt.Errorf("reads item %d is empty; it names a module", i+1))
		default:
			out = append(out, module)
		}
	}
	return out, errors.Join(errs...)
}

func decodeIndex(v any) (map[string]IndexEntry, []error) {
	entries, isObject := v.(map[string]any)
	if !isObject {
		return nil, []error{fmt.Errorf("index must be an object, not %s", jsonvalue.Kind(v))}
	}
	index := map[string]IndexEntry{}
	var errs []error
	for _, p := range sortedKeys(entries) {
		e, err := decodeEntry(p, entries[p])
		if err != nil {
			errs = append(errs, fmt.Errorf("index entry %q: %w", p, err))
		}
		index[p] = e
	}
	for _, p := range sortedKeys(index) {
		for _, err := range []error{checkOutside(index, p), checkParent(index, p)} {
			if err != nil {
				errs = append(errs, fmt.Errorf("index entry %q: %w", p, err))
			}
		}
	}
	return index, errs
}

// decodeEntry decodes one entry, leaving out a parent it refuses, so checkParent names no
// mistake twice.
func decodeEntry(p string, v any) (IndexEntry, error) {
	var e IndexEntry
	errs := []error{checkCategoryPath(p)}
	fields, isObject := v.(map[string]any)
	if !isObject {
		errs = append(errs, fmt.Errorf("an entry must be an object, not %s", jsonvalue.Kind(v)))
		return e, errors.Join(errs...)
	}
	if _, has := fields["paths"]; !has {
		errs = append(errs, fmt.Errorf("paths is missing"))
	}
	for _, k := range sortedKeys(fields) {
		switch k {
		case "parent":
			parent, err := decodeParent(fields[k])
			e.Parent = parent
			errs = append(errs, err)
		case "paths":
			var err error
			e.Paths, err = decodePaths(fields[k])
			errs = append(errs, err)
		default:
			errs = append(errs, fmt.Errorf("unknown key %q; an entry holds parent and paths", k))
		}
	}
	return e, errors.Join(errs...)
}

func decodeParent(v any) (string, error) {
	parent, isString := v.(string)
	if !isString {
		return "", fmt.Errorf("parent must be a string, not %s", jsonvalue.Kind(v))
	}
	if err := grammar.CheckIdentifier(parent); err != nil {
		return "", fmt.Errorf("parent %w", err)
	}
	return parent, nil
}

// decodePaths decodes a list of paths below a category, "" being the category itself.
func decodePaths(v any) ([]string, error) {
	items, isList := v.([]any)
	if !isList {
		return nil, fmt.Errorf("paths must be a list, not %s", jsonvalue.Kind(v))
	}
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	var errs []error
	for i, item := range items {
		q, isString := item.(string)
		if !isString {
			errs = append(errs, fmt.Errorf("paths item %d must be a string, not %s", i+1, jsonvalue.Kind(item)))
			continue
		}
		if q != "" {
			if err := checkCategoryPath(q); err != nil {
				errs = append(errs, fmt.Errorf("paths item %d: %w", i+1, err))
				continue
			}
		}
		if seen[q] {
			errs = append(errs, fmt.Errorf("paths item %d repeats %q", i+1, q))
			continue
		}
		seen[q] = true
		out = append(out, q)
	}
	return out, errors.Join(errs...)
}

// checkOutside refuses an entry whose path passes through another entry's category.
func checkOutside(index map[string]IndexEntry, p string) error {
	segs := strings.Split(p, ".")
	for n := 1; n < len(segs); n++ {
		outer := strings.Join(segs[:n], ".")
		if _, isEntry := index[outer]; isEntry {
			return fmt.Errorf("%q is a category, so it holds no other", outer)
		}
	}
	return nil
}

// checkParent refuses a parent naming no entry beside p: a table loads with its parent,
// from one source.
func checkParent(index map[string]IndexEntry, p string) error {
	parent := index[p].Parent
	if parent == "" {
		return nil
	}
	sibling := parent
	if dot := strings.LastIndexByte(p, '.'); dot >= 0 {
		sibling = p[:dot+1] + parent
	}
	if _, ok := index[sibling]; !ok {
		return fmt.Errorf("parent %q names no entry in this entry's folder; a table in an indexed source links only to a parent in the same index, so to link under another source's table, drop this index, and the source loads whole at start", parent)
	}
	return nil
}

func checkCategoryPath(p string) error {
	for _, seg := range strings.Split(p, ".") {
		if err := grammar.CheckIdentifier(seg); err != nil {
			return fmt.Errorf("%q: %w", p, err)
		}
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
