package fejkdata

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// fakeKey is a word a {reword()} call met, in lower case, and where it read the words it drew a
// fake from.
type fakeKey struct {
	at   tree.WordsAt
	word string
}

// rewordDrawOf builds the draw of o, a {reword()} call; built holds the words of each path read
// so far, so calls reading one path share its words.
func rewordDrawOf(t *tree.Template, o *tree.Op, l *tree.LoadFinds, built map[tree.WordsAt]*tree.WordList) (*tree.RewordDraw, error) {
	words, err := wordsOf(t, o.Operands()[1], l, built)
	if err != nil {
		return nil, err
	}
	return tree.NewRewordDraw(&o.Operands()[0], words), nil
}

func drawReword(d *tree.RewordDraw, s *drawstate.State, t *tree.Template, env renderEnv) string {
	return reword(d.Words(), s, env.frames, readField(s, t, env, *d.Text()).text)
}

// wordsOf is the values a's path renders that are made of letters alone, by length.
func wordsOf(t *tree.Template, a tree.Arm, l *tree.LoadFinds, built map[tree.WordsAt]*tree.WordList) (*tree.WordList, error) {
	if a.Kind() == tree.NamedRead {
		return nil, fmt.Errorf("%q is a name, which holds one pick; pass the path it binds, as in {%s(t,/en_US.word)}", a.Head(), builtinfunc.RewordName)
	}
	at := tree.NewWordsAt(t.StartOf(a.Head(), l.Of(t).Refs()), strings.Join(a.Tail(), "."))
	if words, done := built[at]; done {
		return words, nil
	}
	w := valueWalk{finds: l, seen: map[tree.Node]bool{}}
	if err := w.walk(at.Start(), a.Tail(), 1, "", ""); err != nil {
		return nil, fmt.Errorf("its words are read whole at load, and %w", err)
	}
	var letters []weightedText
	for _, v := range w.out {
		if v.text != "" && !strings.ContainsFunc(v.text, func(r rune) bool { return !unicode.IsLetter(r) }) {
			letters = append(letters, v)
		}
	}
	if len(letters) == 0 {
		return nil, fmt.Errorf("%s holds no word made of letters alone", a.Spelling())
	}
	words := tree.NewWordList(newLengthIndex(letters), at)
	built[at] = words
	return words, nil
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
