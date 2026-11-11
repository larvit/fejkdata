package engine

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/render"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

// InputValue reads an inline template as the value load.Compile takes: the JSON object, array, string
// or null it holds, the whitespace around it dropped, or else the input itself as a format string,
// a bare number or boolean among them.
func InputValue(input string) jsonvalue.Value {
	raw, err := jsonvalue.Decode([]byte(input))
	if err != nil {
		return jsonvalue.String(input)
	}
	switch raw.Any().(type) {
	case json.Number, bool:
		return jsonvalue.String(input)
	}
	return raw
}

// Template is an inline template whose load finished, ready to draw.
type Template struct{ n tree.Node }

// LoadTemplate compiles an inline template, a format string or a JSON value, loads the unloaded
// categories it reads, and runs it through the pipeline against root. A failed lazy load
// matches ErrLoad.
func LoadTemplate(root *tree.Folder, input, label string) (Template, error) {
	l := tree.NewLoadFinds()
	n, err := load.Compile(InputValue(input), l)
	if err != nil {
		return Template{}, err
	}
	if err := loadInline(root, n, l, label, false); err != nil {
		return Template{}, noDataNote(root, err)
	}
	return Template{n}, nil
}

// Draw draws t once.
func (t Template) Draw(s *drawstate.State) string { return render.Once(s, t.n) }

// Record is the record t is.
func (t Template) Record() (Record, error) {
	record, err := recordOf(t.n)
	if err != nil {
		return Record{}, fmt.Errorf("an inline record %w", err)
	}
	return Record{template: record}, nil
}

func loadInline(root *tree.Folder, n tree.Node, l *tree.LoadFinds, label string, typedByGo bool) error {
	nodes := inlineNodes(n, label)
	if err := refuseFolderRefs(nodes); err != nil {
		return err
	}
	if err := loadReached(root, nil, tree.NewLoadFinds(), referenced(root, nil, nodes)); err != nil {
		return loadError{err}
	}
	return pipeline{
		nodes:     nodes,
		finds:     l,
		resolve:   func() ([]load.TemplateSite, error) { return load.ResolveInlineTemplates(n, label, root.Children(), l) },
		typedByGo: typedByGo,
	}.run()
}

// refuseFolderRefs refuses each reference in nodes not written from the root, {/x}: an
// inline node sits in no folder.
func refuseFolderRefs(nodes fence.NodeSet) error {
	return nodes(func(label string, m tree.Node) error {
		t, ok := m.(*tree.Template)
		if !ok {
			return nil
		}
		for _, spelling := range load.RefSpellings(t.Tokens()) {
			sigil, rest, err := grammar.RefShape(spelling)
			if err != nil {
				return fmt.Errorf("%s: reference {%s}: %w", label, spelling, err)
			}
			if sigil != "/" {
				return fmt.Errorf("%s: reference {%s}: an inline template has no folder; write {/%s}", label, spelling, rest)
			}
		}
		return nil
	})
}

// LoadStruct compiles the fake tags of the struct type t as one record, loads it against root,
// and proves each column holds only what its field's Go type can. A failed lazy load matches
// ErrLoad.
func LoadStruct(root *tree.Folder, t reflect.Type, label string, tags map[string]jsonvalue.Value) (Record, error) {
	tags["format"] = jsonvalue.String("")
	l := tree.NewLoadFinds()
	n, err := load.CompileStruct(jsonvalue.Object(tags), l)
	if err != nil {
		return Record{}, fmt.Errorf("%s: %w", label, err)
	}
	if err := loadInline(root, n, l, label, true); err != nil {
		return Record{}, noDataNote(root, err)
	}
	record, err := recordOf(n)
	if err != nil {
		return Record{}, fmt.Errorf("%s: %w", label, err)
	}
	if err := fence.CheckFields(label, t, record.Columns()); err != nil {
		return Record{}, err
	}
	return Record{template: record}, nil
}

// FieldDataType is what a field of the Go kind k holds, and false where no fake tag fills that
// kind.
func FieldDataType(k reflect.Kind) (datatype.DataType, bool) {
	kind, holds := fence.FieldKinds[k]
	return kind.Datatype, holds
}
