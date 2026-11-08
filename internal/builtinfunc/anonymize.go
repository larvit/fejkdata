package builtinfunc

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// AnonymizeName names the builtin that package fejkdata renders, since only that package holds the
// words a call's list renders and the fakes a render has drawn.
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
