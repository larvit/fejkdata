package tree_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

const earlyRead = "before LoadFinds.Finish"

func panicOf(read func()) (r any) {
	defer func() { r = recover() }()
	read()
	return nil
}

func unfinishedBinding() *tree.NameBinding {
	t := tree.NewTemplate(tree.TemplateSyntax{Format: "{x as n}"})
	sc := tree.NewNameScope(nil, t)
	sc.Bind(grammar.Token{Bound: "n", BoundRef: "x"}, "", t)
	return sc.Bindings()["n"]
}

func TestAReadOfAnUnfinishedTemplatePanicsNamingThePhase(t *testing.T) {
	unfinished := tree.NewTemplate(tree.TemplateSyntax{Format: "plain"})
	cases := map[string]func(){
		"Columns":            func() { unfinished.Columns() },
		"Compiled":           func() { unfinished.Compiled() },
		"NameScope":          func() { unfinished.NameScope() },
		"OwnScope":           func() { unfinished.OwnScope() },
		"ReadsColumn":        func() { unfinished.ReadsColumn() },
		"Refs":               func() { unfinished.Refs() },
		"Of after Finish":    func() { tree.AfterFinish.Of(unfinished) },
		"Of of another load": func() { tree.NewLoadFinds().Of(unfinished) },
	}
	for name, read := range cases {
		t.Run(name, func(t *testing.T) {
			if r := panicOf(read); !strings.Contains(fmt.Sprint(r), `template "plain" is read `+earlyRead) {
				t.Errorf("recovered %v, want a panic naming the template and the phase", r)
			}
		})
	}
}

func TestAReadOfAnUnfinishedBindingPanicsNamingThePhase(t *testing.T) {
	unfinished := unfinishedBinding()
	cases := map[string]func(){
		"Addressed":            func() { unfinished.Addressed() },
		"Target":               func() { unfinished.Target() },
		"Values":               func() { unfinished.Values() },
		"Binding after Finish": func() { tree.AfterFinish.Binding(unfinished) },
	}
	for name, read := range cases {
		t.Run(name, func(t *testing.T) {
			if r := panicOf(read); !strings.Contains(fmt.Sprint(r), `name "n" is read `+earlyRead) {
				t.Errorf("recovered %v, want a panic naming the name and the phase", r)
			}
		})
	}
}

func TestOfReadsWhatTheLoadHoldsBeforeFinish(t *testing.T) {
	tmpl := tree.NewTemplate(tree.TemplateSyntax{Format: "plain"})
	l := tree.NewLoadFinds()
	sc := tree.NewNameScope(nil, tmpl)
	l.SetNameScope(tmpl, sc)
	if got := l.Of(tmpl).NameScope(); got != sc {
		t.Errorf("Of(t).NameScope() = %p, want the scope set, %p", got, sc)
	}
}

func TestAFinishedTemplateReadsWhatFinishWrote(t *testing.T) {
	tmpl := tree.NewTemplate(tree.TemplateSyntax{Format: "plain"})
	l := tree.NewLoadFinds()
	sc := tree.NewNameScope(nil, tmpl)
	l.SetNameScope(tmpl, sc)
	l.Finish()
	if got := tmpl.NameScope(); got != sc {
		t.Errorf("NameScope() = %p, want the scope Finish wrote, %p", got, sc)
	}
}
