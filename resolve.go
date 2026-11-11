package fejkdata

import (
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]tree.Node, l *tree.LoadFinds) ([]load.TemplateSite, error) {
	var ts []load.TemplateSite
	for _, s := range sites {
		if err := load.EachNode(s.n, s.path, func(label string, n tree.Node) error {
			if t, isTemplate := n.(*tree.Template); isTemplate {
				ts = append(ts, load.TemplateSite{Template: t, Folder: s.dir, Label: label})
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if err := load.CompileTemplates(ts, root, l); err != nil {
		return nil, err
	}
	return ts, nil
}
