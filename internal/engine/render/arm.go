package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// drawToken prints one draw of d in t; emitted is the output so far.
func drawToken(s *drawstate.State, t *tree.Template, env renderEnv, d tree.TokenDraw, emitted string) string {
	switch d := d.(type) {
	case *tree.ValueDraw:
		return drawValue(d, s, env)
	case *tree.AlphabetDraw:
		return drawAlphabet(d, s, env, emitted)
	case *tree.RewordDraw:
		return drawReword(d, s, t, env)
	}
	panic(invariant.Broken("no draw for %T", d))
}
