package render

import (
	"strings"
	"unicode"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// fakeKey is a word a {reword()} call met, in lower case, and where it read the words it drew a
// fake from.
type fakeKey struct {
	at   tree.WordsAt
	word string
}

func drawReword(d *tree.RewordDraw, s *drawstate.State, t *tree.Template, env renderEnv) string {
	return reword(d.Words(), s, env.frames, readField(s, t, env, *d.Text()).text)
}

// reword rewrites each run of letters in text as a word of its length, or of the nearest length
// x holds, one fake per word in the render; every other rune stays.
func reword(x *tree.WordList, s *drawstate.State, frames *frameStack, text string) string {
	var b strings.Builder
	rs := []rune(text)
	for i := 0; i < len(rs); {
		j := i
		for j < len(rs) && unicode.IsLetter(rs[j]) {
			j++
		}
		if j == i {
			b.WriteRune(rs[i])
			i++
			continue
		}
		writeInCase(&b, fakeWord(x, s, frames, rs[i:j]), rs[i:j])
		i = j
	}
	return b.String()
}

func fakeWord(x *tree.WordList, s *drawstate.State, frames *frameStack, word []rune) string {
	key := fakeKey{x.At(), strings.ToLower(string(word))}
	if fake, drawn := frames.fakes[key]; drawn {
		return fake
	}
	if frames.fakes == nil {
		frames.fakes = map[fakeKey]string{}
	}
	from, to := x.Index().Nearest(len(word))
	fake := drawIn(x.Index(), s, from, to)
	frames.fakes[key] = fake
	return fake
}

// writeInCase writes fake in the case of word's letter at each place, and a letter past word's
// end in the case of its last letter.
func writeInCase(b *strings.Builder, fake string, word []rune) {
	k := 0
	for _, r := range fake {
		if unicode.IsUpper(word[min(k, len(word)-1)]) {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
		k++
	}
}
