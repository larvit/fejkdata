package render

import (
	"slices"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

func fallbackAt(f tree.Fallback, s *drawstate.State, env renderEnv) string {
	if f.Read() == nil {
		return f.Text()
	}
	return readName(s, env, *f.Read()).text
}

func drawValue(v *tree.ValueDraw, s *drawstate.State, env renderEnv) string {
	return strings.Join(drawValues(v, s, env), v.Sep())
}

// drawValues draws count distinct passing values, reading the clause's names in env. Where the
// strings it reads leave fewer values than count, the fallback prints in place of each value missing.
func drawValues(v *tree.ValueDraw, s *drawstate.State, env renderEnv) []string {
	from, to := v.Index().Span(lengthEndAt(v.Lo(), s, env), lengthEndAt(v.Hi(), s, env))
	if to-from < v.Count() {
		panic(invariant.Broken("a clause the load proved leaves %d values leaves %d", v.Count(), to-from))
	}
	drawn := holding(v, s, env, from, to) // ascending
	out := make([]string, v.Count())
	for k := range out {
		if len(drawn) == to-from {
			out[k] = fallbackAt(v.Fallback(), s, env)
			continue
		}
		i := drawIndex(v.Index(), s, from, to, drawn)
		at, _ := slices.BinarySearch(drawn, i)
		drawn = slices.Insert(drawn, at, i)
		out[k] = v.Index().Texts()[i]
	}
	return out
}

// holding is the indexes, ascending, of the texts in v.Index().Texts()[from:to] that equal a not string, or
// hold a not containing string, that v reads in env.
func holding(v *tree.ValueDraw, s *drawstate.State, env renderEnv, from, to int) []int {
	not, containing := readClauseStrings(v.Strings(), s, env)
	if len(not) == 0 && len(containing) == 0 {
		return nil
	}
	var out []int
	for i := from; i < to; i++ {
		if slices.Contains(not, v.Index().Texts()[i]) || avoid.HoldsAny(v.Index().Texts()[i], containing) {
			out = append(out, i)
		}
	}
	return out
}

// readClauseStrings reads r's strings in env: those a value must not equal, and those a value
// must not hold, leaving out an empty one, which every value holds.
func readClauseStrings(r tree.ReadStrings, s *drawstate.State, env renderEnv) (not, containing []string) {
	for _, a := range r.Not() {
		not = append(not, readName(s, env, *a).text)
	}
	for _, a := range r.NotContaining() {
		if text := readName(s, env, *a).text; text != "" {
			containing = append(containing, text)
		}
	}
	return not, containing
}

func lengthEndAt(e tree.LengthEnd, s *drawstate.State, env renderEnv) int {
	if e.Read() == nil {
		return e.N()
	}
	text := readName(s, env, *e.Read()).text
	v, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		panic(invariant.Broken("a length bound read %q, which a load proved a number", text))
	}
	return int(v)
}
