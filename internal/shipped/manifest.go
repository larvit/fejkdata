package shipped

import (
	"encoding/json"
	"slices"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine"
)

// indexedModule is the lines of one module's index, one category a line, and the modules its
// categories read directly, each by its place in Data's namespaces.
type indexedModule struct {
	lines  []string
	direct []int
}

func (d *Data) indexModules() ([]indexedModule, error) {
	modules := make([]indexedModule, len(d.namespaces))
	for _, s := range engine.Index(&d.root) {
		i := d.holder(s.Path)
		line, err := json.Marshal(s.Entry)
		if err != nil {
			return nil, err
		}
		key, _ := json.Marshal(s.Path)
		modules[i].lines = append(modules[i].lines, "\n\t\t"+string(key)+": "+string(line))
		for _, c := range s.Needs {
			if j := d.holder(c); j != i && !slices.Contains(modules[i].direct, j) {
				modules[i].direct = append(modules[i].direct, j)
			}
		}
	}
	return modules, nil
}

// Manifests is each module's manifest, read from the whole tree: its index and its namespace.
func (d *Data) Manifests() ([]File, error) {
	modules, err := d.indexModules()
	if err != nil {
		return nil, err
	}
	out := make([]File, len(modules))
	for i, m := range modules {
		namespace, _ := json.Marshal(d.namespaces[i])
		name := "data/" + strings.ReplaceAll(d.namespaces[i], ".", "/") + "/" + datafiles.ManifestFile
		out[i] = File{name, []byte("{\n\t\"index\": {" + strings.Join(m.lines, ",") + "\n\t},\n\t\"namespace\": " + string(namespace) + "\n}\n")}
	}
	return out, nil
}

// Reads is, by a module's namespace, the namespace of every other module it reads, directly or
// through another, sorted.
func (d *Data) Reads() (map[string][]string, error) {
	modules, err := d.indexModules()
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for i, n := range d.namespaces {
		out[n] = d.readsOf(modules, i)
	}
	return out, nil
}

func (d *Data) readsOf(modules []indexedModule, i int) []string {
	seen := map[int]bool{i: true}
	queue := slices.Clone(modules[i].direct)
	var out []string
	for ; len(queue) > 0; queue = queue[1:] {
		j := queue[0]
		if seen[j] {
			continue
		}
		seen[j] = true
		out = append(out, d.namespaces[j])
		queue = append(queue, modules[j].direct...)
	}
	sort.Strings(out)
	return out
}
