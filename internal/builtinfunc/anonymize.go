package builtinfunc

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// AnonymizeName is the builtin package fejkdata renders, since only it holds the words its list
// renders and the fakes a render has drawn.
const AnonymizeName = "anonymize"

func pathArgs(a []string) error {
	for _, arg := range a {
		if _, _, isCall := grammar.FuncCall(arg); isCall {
			return fmt.Errorf("%s is a call; each argument is a path, as in {%s(t,/en_US.word)}", arg, AnonymizeName)
		}
	}
	return nil
}

func renderedByFejkdata([]string) Call {
	return func(*drawstate.State, string, []string) string {
		panic(invariant.Broken("{%s()} renders in package fejkdata, which holds its words", AnonymizeName))
	}
}
