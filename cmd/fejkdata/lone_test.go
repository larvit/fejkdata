package main

import (
	"testing"

	"github.com/larvit/fejkdata/internal/grammar"
)

// grammarLoneReference is the path a format reads when the grammar parses it as one path
// read of one reference from the root, and nothing else.
func grammarLoneReference(text string) (string, bool) {
	toks, err := grammar.ParseFormat(text)
	if err != nil || len(toks) != 1 || toks[0].Kind != grammar.PathRead || len(toks[0].Arms) != 1 {
		return "", false
	}
	sigil, path, err := grammar.RefShape(toks[0].Arms[0])
	if err != nil || sigil != "/" {
		return "", false
	}
	return path, true
}

func TestLoneReferenceAgreesWithTheGrammar(t *testing.T) {
	for _, text := range []string{
		"{/sv_SE.person}",
		"{/misc.currency[US Dollar (Next day)]}",
		"{/t[a as b].c}",
		"{/t[f(x)]}",
		"{/a}{/b}",
		"{/x}}",
		"{/}",
		"{/.x}",
		" {/x}",
		"{/a|/b}",
		"{/x as p}",
		"{//x}",
		"{/x(1)}",
		"{.x}",
		"x",
	} {
		path, lone := loneReference(text)
		want, wantLone := grammarLoneReference(text)
		if lone != wantLone || path != want {
			t.Errorf("loneReference(%q) = %q, %v; the grammar reads %q, %v", text, path, lone, want, wantLone)
		}
	}
}
