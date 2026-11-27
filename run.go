package fejkdata

import (
	"errors"
	"fmt"
	"hash/fnv"
	"path"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine"
)

// Emit is one template of a run: the path of a category, and how many records of it the run writes.
type Emit struct {
	Template string
	Count    int
}

// ErrRun marks a run that [FakeRun] refuses before it renders a record, for what its emits name.
var ErrRun = errors.New("fejkdata: the run's templates cannot be written")

type runError struct{ error }

func (e runError) Unwrap() error      { return e.error }
func (runError) Is(target error) bool { return target == ErrRun }

// FakeRun writes several linked templates in one run. A template is a category, rendered as
// [Generator.FakeRecord] renders it. It reads the rows another template of the run wrote through
// {/run.<template>}, a table of that template's columns whose format is its first column:
// {/run.customers as c} keeps one customer, and {c.id} and {c.name} read that row. The template
// read is the parent, and the one reading it the child.
//
// FakeRun calls write once per emit, and each call of fake renders that template's next record,
// which the run keeps as a row for its children. The emits come in the order given, save that a
// parent moves ahead of its first child. opts are those of [New]. A template draws from the seed
// and its own name, so its records stay the same whatever else the run names, and a
// {seq()} counts within one template.
//
// The error matches [ErrRun], before the first write, where a template reads the rows of one the
// run doesn't name, of one with count 0 while its own is above 0, or of one that reads its rows
// back; where an emit names no category, a category twice or a count below 0; and where the data
// holds a root entry named run. It matches [ErrLoad] where the data fails to load, which a child
// reading a column its parent lacks does once the parent is written. An error write returns is
// returned as it is.
func FakeRun(emits []Emit, write func(e Emit, fake func() (*Record, error)) error, opts ...Option) error {
	var c config
	for _, opt := range opts {
		opt(&c)
	}
	if err := c.misuse(); err != nil {
		return err
	}
	compiled, unloaded, err := engine.References(c.sources)
	if err != nil {
		return fmt.Errorf("fejkdata: %w", err)
	}
	plan, err := planRun(emits, compiled, unloaded)
	if err != nil {
		return runError{fmt.Errorf("fejkdata: %w", err)}
	}
	return plan.fake(c, write)
}

// runPlan is a run FakeRun can write.
type runPlan struct {
	emits []Emit // in the order to write them
	// needs is, by category, what must be written before the category loads: each template whose
	// rows it reads, itself or through what it references, and each /run path naming none.
	needs map[string]map[string]bool
	// unnamed is, by /run path naming no template, the first category spelling it.
	unnamed map[string]string
}

func planRun(emits []Emit, compiled []engine.CategoryReads, unloaded []string) (runPlan, error) {
	if err := checkEmits(emits, compiled, unloaded); err != nil {
		return runPlan{}, err
	}
	p := runPlan{needs: map[string]map[string]bool{}, unnamed: map[string]string{}}
	p.readNeeds(emits, compiled)
	for _, e := range emits {
		if err := p.refusal(e, emits); err != nil {
			return runPlan{}, err
		}
	}
	placed := map[string]bool{}
	for _, e := range emits {
		p.place(e, emits, placed)
	}
	return p, nil
}

// checkEmits refuses a run for what its emits name, whatever the templates read.
func checkEmits(emits []Emit, compiled []engine.CategoryReads, unloaded []string) error {
	if len(emits) == 0 {
		return errors.New("the run names no template")
	}
	categories := map[string]bool{}
	for _, c := range compiled {
		unloaded = append(unloaded, c.Path)
	}
	for _, category := range unloaded {
		if category == "run" || strings.HasPrefix(category, "run.") {
			return fmt.Errorf(`the data holds %s, and a run keeps the root name "run" for the rows its templates read; rename it`, category)
		}
		categories[category] = true
	}
	named := map[string]bool{}
	for _, e := range emits {
		if err := checkEmit(e, categories, named); err != nil {
			return err
		}
		named[e.Template] = true
	}
	return nil
}

func checkEmit(e Emit, categories, named map[string]bool) error {
	switch {
	case e.Count < 0:
		return fmt.Errorf("template %s has count %d, and a count is 0 or more", e.Template, e.Count)
	case named[e.Template]:
		return fmt.Errorf("template %s is named twice", e.Template)
	case !categories[e.Template]:
		return fmt.Errorf("template %s is no category of the data", e.Template)
	}
	return nil
}

// readNeeds fills needs and unnamed from what each category spells, then gives each category the
// needs of every category it references and of every template it reads.
func (p *runPlan) readNeeds(emits []Emit, compiled []engine.CategoryReads) {
	isCategory := map[string]bool{}
	for _, c := range compiled {
		isCategory[c.Path] = true
	}
	references := map[string][]string{}
	for _, c := range compiled {
		p.needs[c.Path] = map[string]bool{}
		for _, segs := range c.Reads {
			if len(segs) > 0 && segs[0] == "run" {
				references[c.Path] = append(references[c.Path], p.readRows(c.Path, emits, segs[1:])...)
			} else if category := longestPath(segs, isCategory); category != "" {
				references[c.Path] = append(references[c.Path], category)
			}
		}
	}
	for spread := true; spread; {
		spread = false
		for category, read := range references {
			spread = p.inherit(category, read) || spread
		}
	}
}

// readRows records that category reads /run.<rest>, and returns the template whose rows that reads,
// if the run names it.
func (p *runPlan) readRows(category string, emits []Emit, rest []string) []string {
	named := map[string]bool{}
	for _, e := range emits {
		named[e.Template] = true
	}
	if template := longestPath(rest, named); template != "" {
		p.needs[category][template] = true
		return []string{template}
	}
	read := "/run." + strings.Join(rest, ".")
	p.needs[category][read] = true
	if _, spelled := p.unnamed[read]; !spelled {
		p.unnamed[read] = category
	}
	return nil
}

// inherit adds to category's needs those of each category in read, and reports whether it added any.
func (p *runPlan) inherit(category string, read []string) bool {
	added := false
	for _, from := range read {
		for need := range p.needs[from] {
			if !p.needs[category][need] {
				p.needs[category][need], added = true, true
			}
		}
	}
	return added
}

// longestPath is the longest dot path in paths that segs starts with, "" where none is.
func longestPath(segs []string, paths map[string]bool) string {
	for i := len(segs); i > 0; i-- {
		if p := strings.Join(segs[:i], "."); paths[p] {
			return p
		}
	}
	return ""
}

// refusal is why the run can't write e, nil where it can.
func (p runPlan) refusal(e Emit, emits []Emit) error {
	needs := p.needs[e.Template]
	var unnamed []string
	for need := range needs {
		if strings.HasPrefix(need, "/") {
			unnamed = append(unnamed, need)
		}
	}
	if len(unnamed) > 0 {
		sort.Strings(unnamed)
		return fmt.Errorf("%s reads %s, a template the run doesn't name", p.unnamed[unnamed[0]], unnamed[0])
	}
	for _, parent := range emits {
		if err := p.parentRefusal(e, parent); err != nil {
			return err
		}
	}
	if needs[e.Template] {
		return fmt.Errorf("%s reads the rows of %s, its own; a template reads only another's rows", e.Template, e.Template)
	}
	return nil
}

// parentRefusal is why the run can't write e after parent, nil where it can or where e reads no row of it.
func (p runPlan) parentRefusal(e, parent Emit) error {
	switch {
	case parent.Template == e.Template || !p.needs[e.Template][parent.Template]:
		return nil
	case p.needs[parent.Template][e.Template]:
		return fmt.Errorf("%s reads the rows of %s, which reads the rows of %s: a cycle, so neither can be written first", e.Template, parent.Template, e.Template)
	case e.Count > 0 && parent.Count == 0:
		return fmt.Errorf("%s reads the rows of %s, whose count is 0; give %s a count above 0, or %s the count 0", e.Template, parent.Template, parent.Template, e.Template)
	}
	return nil
}

// place appends e to the plan's emits, after each template it needs.
func (p *runPlan) place(e Emit, emits []Emit, placed map[string]bool) {
	if placed[e.Template] {
		return
	}
	placed[e.Template] = true
	for _, parent := range emits {
		if p.needs[e.Template][parent.Template] {
			p.place(parent, emits, placed)
		}
	}
	p.emits = append(p.emits, e)
}

// fake writes each template in turn, loading the data once per template: over the rows written so
// far, and without the categories that read rows not yet written.
func (p runPlan) fake(c config, write func(e Emit, fake func() (*Record, error)) error) error {
	tables := map[string]string{}
	for _, e := range p.emits {
		g, err := p.generator(c, e.Template, tables)
		if err != nil {
			return err
		}
		rows := runRows{g: g, template: e.Template, kept: p.isRead(e.Template)}
		if err := write(e, rows.fake); err != nil {
			return err
		}
		if err := rows.keepInto(tables); err != nil {
			return err
		}
	}
	return nil
}

func (p runPlan) isRead(template string) bool {
	for _, needs := range p.needs {
		if needs[template] {
			return true
		}
	}
	return false
}

// generator loads the data for rendering template: each table of tables at /run.<its template>,
// and an empty category in place of each one that needs rows not yet in tables.
func (p runPlan) generator(c config, template string, tables map[string]string) (*Generator, error) {
	files := map[string]string{}
	for written, tsv := range tables {
		file := "run/" + strings.ReplaceAll(written, ".", "/")
		firstColumn, _, _ := strings.Cut(tsv[:strings.Index(tsv, "\n")], "\t")
		files[file+".json"] = fmt.Sprintf(`{"format":"{%s}","rows":"%s.tsv"}`, firstColumn, path.Base(file))
		files[file+".tsv"] = tsv
	}
	for category, needs := range p.needs {
		for need := range needs {
			if _, written := tables[need]; !written {
				files[strings.ReplaceAll(category, ".", "/")+".json"] = `{"format":""}`
			}
		}
	}
	root, err := engine.Load(append(c.sources[:len(c.sources):len(c.sources)], datafiles.Files(files)))
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	h := fnv.New64a()
	h.Write([]byte(template))
	return &Generator{drawState: drawstate.New(c.seed^h.Sum64(), c.today), root: root}, nil
}

// runRows renders one template's records, and keeps each as a table row where a category reads them.
type runRows struct {
	g        *Generator
	template string
	kept     bool
	tsv      strings.Builder // the header line, then a line per record
}

func (r *runRows) fake() (*Record, error) {
	record, err := r.g.FakeRecord(r.template)
	if err != nil || !r.kept {
		return record, err
	}
	if r.tsv.Len() == 0 {
		r.tsv.WriteString(record.TSVHeader() + "\n")
	}
	r.tsv.WriteString(record.TSVLine() + "\n")
	return record, nil
}

// keepInto puts the rows into tables. A table holds at least one row, so a template nobody rendered
// renders one here: only children of count 0 load over it, for their columns.
func (r *runRows) keepInto(tables map[string]string) error {
	if !r.kept {
		return nil
	}
	if r.tsv.Len() == 0 {
		if _, err := r.fake(); err != nil {
			return err
		}
	}
	tables[r.template] = r.tsv.String()
	return nil
}
