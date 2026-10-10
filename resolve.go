package fejkdata

import (
	"errors"
	"fmt"
	"strings"
)

// templateSite is a template to resolve, with its folder and category, both empty for an inline
// template, and eachNode's label for it.
type templateSite struct {
	t        *template
	folder   []string
	label    string
	category *categorySite
}

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]node) error {
	var ts []templateSite
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(label string, n node) error {
			if t, isTemplate := n.(*template); isTemplate {
				ts = append(ts, templateSite{t: t, folder: s.dir, label: label, category: &s})
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return resolveTemplates(ts, root)
}

func resolveInlineTemplates(nodes nodeSet, root map[string]node) error {
	var ts []templateSite
	if err := nodes(func(label string, n node) error {
		if t, isTemplate := n.(*template); isTemplate {
			ts = append(ts, templateSite{t: t, label: label})
		}
		return nil
	}); err != nil {
		return err
	}
	return resolveTemplates(ts, root)
}

// resolvedNames is what resolveTemplates resolved about the names of a set of templates: each binding's
// target, and the keys its reads address.
type resolvedNames struct {
	targets   map[*nameBinding]nameTarget
	addressed map[*nameBinding]map[pickKey]string
}

// resolveTemplates resolves ts in steps, each over every template before the next starts. The
// names' targets, keys and reads pass to later steps as arguments, and the bindings take their
// targets and keys only once every check has passed.
func resolveTemplates(ts []templateSite, root map[string]node) error {
	for _, s := range ts {
		refs, err := s.t.resolveRefs(s.folder, s.label, root)
		if err != nil {
			return s.withReads(err)
		}
		s.t.refs = refs
	}
	targets := nameTargets(ts)
	for _, s := range ts {
		compiled, err := compileFormat(s.t, targets)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		s.t.compiled = compiled
	}
	names := resolvedNames{targets: targets, addressed: addressedKeys(ts, targets)}
	for _, s := range ts {
		s.t.readsColumn = columnReadOf(s.t, targets)
	}
	for _, s := range ts {
		if err := checkNameReads(s.label, s.t, names); err != nil {
			return err
		}
	}
	for _, s := range ts {
		if err := checkCalcNames(s.label, s.t); err != nil {
			return err
		}
	}
	for b, target := range targets {
		b.target, b.addressed = target, names.addressed[b]
	}
	return nil
}

// withReads adds to err, where a reference of a category names nothing, the modules the
// category's manifest names it reading by default, or that it names none.
func (s templateSite) withReads(err error) error {
	var ne noEntry
	if s.category == nil || !errors.As(err, &ne) {
		return err
	}
	if len(s.category.reads) == 0 {
		return fmt.Errorf("%w, and the module holding %s names no module it reads by default", err, s.category.path)
	}
	return fmt.Errorf("%w; the module holding %s reads by default %s", err, s.category.path, strings.Join(s.category.reads, ", "))
}
