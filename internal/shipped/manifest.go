package shipped

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// ManifestFile is the name of the manifest of the module mounted under namespace.
func ManifestFile(namespace string) string {
	return "data/" + strings.ReplaceAll(namespace, ".", "/") + "/" + datafiles.ManifestFile
}

// moduleIndex is, for each module, the lines of its manifest's index, one category a line, and
// the modules its categories read directly.
type moduleIndex struct {
	lines  [][]string
	direct [][]int
}

func indexModules(root *tree.Folder, namespaces []string) (moduleIndex, error) {
	index := moduleIndex{lines: make([][]string, len(namespaces)), direct: make([][]int, len(namespaces))}
	for _, s := range engine.Index(root) {
		i := holder(namespaces, s.Path)
		if i < 0 {
			return index, fmt.Errorf("no shipped module holds %s", s.Path)
		}
		line, err := json.Marshal(s.Entry)
		if err != nil {
			return index, err
		}
		key, _ := json.Marshal(s.Path)
		index.lines[i] = append(index.lines[i], "\n\t\t"+string(key)+": "+string(line))
		for _, c := range s.Needs {
			if j := holder(namespaces, c); j != i && !slices.Contains(index.direct[i], j) {
				index.direct[i] = append(index.direct[i], j)
			}
		}
	}
	return index, nil
}

// Manifests is each module's manifest, read from root's whole tree: its index and its namespace.
func Manifests(root *tree.Folder, namespaces []string) ([]File, error) {
	index, err := indexModules(root, namespaces)
	if err != nil {
		return nil, err
	}
	out := make([]File, len(namespaces))
	for i, n := range namespaces {
		namespace, _ := json.Marshal(n)
		out[i] = File{ManifestFile(n), []byte("{\n\t\"index\": {" + strings.Join(index.lines[i], ",") + "\n\t},\n\t\"namespace\": " + string(namespace) + "\n}\n")}
	}
	return out, nil
}

// Reads is, for each module, the namespace of every other module it reads, directly or through
// another, sorted.
func Reads(root *tree.Folder, namespaces []string) ([][]string, error) {
	index, err := indexModules(root, namespaces)
	if err != nil {
		return nil, err
	}
	out := make([][]string, len(namespaces))
	for i := range namespaces {
		out[i] = index.readsOf(i, namespaces)
	}
	return out, nil
}

func (index moduleIndex) readsOf(i int, namespaces []string) []string {
	seen := map[int]bool{i: true}
	queue := slices.Clone(index.direct[i])
	var out []string
	for ; len(queue) > 0; queue = queue[1:] {
		j := queue[0]
		if seen[j] {
			continue
		}
		seen[j] = true
		out = append(out, namespaces[j])
		queue = append(queue, index.direct[j]...)
	}
	sort.Strings(out)
	return out
}
