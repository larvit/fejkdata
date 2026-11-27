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

// ErrRun marks a run that [FakeRun] refuses before it loads the data: its emits, or what its
// templates read of each other, can't make a run. FakeRun's doc lists each case.
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
// FakeRun calls write once per emit, and each call of fake renders that template's next record.
// A child reads its parent's first Count records, the ones fake returns. FakeRun renders them once
// before the first write as well. The emits come in the order given, save that a parent moves ahead
// of its first child. opts are those of [New]. A template draws from the seed and its own name, so
// its records stay the same whatever else the run names, and a {seq()} counts within one template.
//
// FakeRun returns an error before the first write in each of these cases. The error matches [ErrRun]
// where:
//   - emits is empty, or an emit names no category, a category twice or a count below 0;
//   - a template reads the rows of a template the run doesn't name;
//   - a template with a count above 0 reads the rows of one with count 0;
//   - a template reads its own rows, or two templates read each other's;
//   - the data holds a root entry named run.
//
// The error matches [ErrLoad] where the data fails to load, as it does where a child reads a column
// its parent lacks, and it is the error of [New] where opts are refused. After that, FakeRun returns
// the first error write returns, as it is.
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
	categories, err := runCategories(compiled, unloaded)
	if err != nil {
		return err
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

// runCategories is every category of the data by its path, refused where one sits at the root name "run".
func runCategories(compiled []engine.CategoryReads, unloaded []string) (map[string]bool, error) {
	categories := map[string]bool{}
	for _, c := range compiled {
		unloaded = append(unloaded, c.Path)
	}
	for _, category := range unloaded {
		if category == "run" || strings.HasPrefix(category, "run.") {
			return nil, fmt.Errorf(`the data holds %s, and a run keeps the root name "run" for the rows its templates read; rename it`, category)
		}
		categories[category] = true
	}
	return categories, nil
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
		p.needs[c.Path] = map[string]bool{}
	}
	references := map[string][]string{}
	for _, c := range compiled {
		references[c.Path] = p.referenced(c, emits, isCategory)
	}
	for spread := true; spread; {
		spread = false
		for category, read := range references {
			spread = p.inherit(category, read) || spread
		}
	}
}

// referenced is each category c references and each template whose rows it reads, as c spells them.
func (p *runPlan) referenced(c engine.CategoryReads, emits []Emit, isCategory map[string]bool) []string {
	var read []string
	for _, segs := range c.Reads {
		if len(segs) > 0 && segs[0] == "run" {
			read = append(read, p.readRows(c.Path, emits, segs[1:])...)
		} else if category := longestPath(segs, isCategory); category != "" {
			read = append(read, category)
		}
	}
	return read
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
		return fmt.Errorf("%s reads %s, and the run doesn't name the template those rows belong to", p.unnamed[unnamed[0]], unnamed[0])
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

// fake writes each template in turn, loading the data for it over the tables dryLoad returns.
func (p runPlan) fake(c config, write func(e Emit, fake func() (*Record, error)) error) error {
	tables, err := p.dryLoad(c)
	if err != nil {
		return err
	}
	for _, e := range p.emits {
		g, err := p.generator(c, e.Template, tables)
		if err != nil {
			return err
		}
		if err := write(e, func() (*Record, error) { return g.FakeRecord(e.Template) }); err != nil {
			return err
		}
	}
	return nil
}

// dryLoad loads each template in turn and renders the rows of each one a child reads, so data that
// fails to load fails before the first record is written. It returns those rows as tables, by template.
func (p runPlan) dryLoad(c config) (map[string]string, error) {
	tables := map[string]string{}
	for _, e := range p.emits {
		g, err := p.generator(c, e.Template, tables)
		if err != nil {
			return nil, err
		}
		if !p.isRead(e.Template) {
			continue
		}
		if err := keepInto(tables, g, e); err != nil {
			return nil, err
		}
	}
	return tables, nil
}

// isRead reports whether a template of the run reads template's rows.
func (p runPlan) isRead(template string) bool {
	for _, e := range p.emits {
		if p.needs[e.Template][template] {
			return true
		}
	}
	return false
}

// generator loads the data for rendering template: each table of tables that template needs, at
// /run.<its template>, and an empty category in place of each one that needs any other rows.
func (p runPlan) generator(c config, template string, tables map[string]string) (*Generator, error) {
	files := map[string]string{}
	for parent := range p.needs[template] {
		tsv := tables[parent]
		file := "run/" + strings.ReplaceAll(parent, ".", "/")
		firstColumn, _, _ := strings.Cut(tsv[:strings.Index(tsv, "\n")], "\t")
		files[file+".json"] = fmt.Sprintf(`{"format":"{%s}","rows":"%s.tsv"}`, firstColumn, path.Base(file))
		files[file+".tsv"] = tsv
	}
	for category, needs := range p.needs {
		for need := range needs {
			if !p.needs[template][need] {
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

// keepInto renders e's Count records from g and puts them into tables as the rows a child reads. A
// template draws from the seed and its own name, so they are the records write gets. A table can't
// be empty, so at count 0 keepInto renders one record.
func keepInto(tables map[string]string, g *Generator, e Emit) error {
	var tsv strings.Builder
	for i := 0; i < max(e.Count, 1); i++ {
		record, err := g.FakeRecord(e.Template)
		if err != nil {
			return err
		}
		if i == 0 {
			tsv.WriteString(record.TSVHeader() + "\n")
		}
		tsv.WriteString(record.TSVLine() + "\n")
	}
	tables[e.Template] = tsv.String()
	return nil
}
