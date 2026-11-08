package fejkdata

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
)

// fakeKey is a word an {anonymize()} call met, in lower case, and the words it drew a fake from.
type fakeKey struct {
	words *lengthIndex
	word  string
}

// wordsAt is where an {anonymize()} call reads its words: the node its path starts at, and the rest.
type wordsAt struct {
	start node
	tail  string
}

// anonymizeWords builds the words each {anonymize()} call in ts draws from, once per path they
// are read at, so calls reading one list keep one fake per word.
func anonymizeWords(ts []templateSite, targets map[*nameBinding]nameTarget) error {
	built := map[wordsAt]*lengthIndex{}
	for _, s := range ts {
		for i := range s.t.compiled.ops {
			o := &s.t.compiled.ops[i]
			if o.Kind != grammar.BuiltinCall || o.Fn != builtinfunc.AnonymizeName {
				continue
			}
			words, err := s.t.wordsOf(o.operands[1], targets, built)
			if err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.label, o.Body, err)
			}
			o.words = words
		}
	}
	return nil
}

// wordsOf is the values a's path renders that are made of letters alone, by length.
func (t *template) wordsOf(a arm, targets map[*nameBinding]nameTarget, built map[wordsAt]*lengthIndex) (*lengthIndex, error) {
	if a.kind == namedRead {
		return nil, fmt.Errorf("%q is a name, which holds one pick; read the words whole, as in {%s(t,/en_US.word)}", a.head, builtinfunc.AnonymizeName)
	}
	at := wordsAt{t.startOf(a.head), strings.Join(a.tail, ".")}
	if words, done := built[at]; done {
		return words, nil
	}
	w := valueWalk{targets: targets, seen: map[node]bool{}}
	if err := w.walk(at.start, a.tail, 1, "", ""); err != nil {
		return nil, fmt.Errorf("its words are read whole at load, and %w", err)
	}
	var letters []filterValue
	for _, v := range w.out {
		if v.text != "" && !strings.ContainsFunc(v.text, func(r rune) bool { return !unicode.IsLetter(r) }) {
			letters = append(letters, v)
		}
	}
	if len(letters) == 0 {
		return nil, fmt.Errorf("%s holds no word made of letters alone", a.spelling)
	}
	words := newLengthIndex(letters)
	built[at] = &words
	return &words, nil
}

// anonymize rewrites each word of text, a run of letters, as a word of its length or the nearest
// length x holds, one fake per word in the render; every other rune stays.
func (x *lengthIndex) anonymize(s *drawstate.State, frames *frameStack, text string) string {
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
		writeInCase(&b, x.fake(s, frames, rs[i:j]), rs[i:j])
		i = j
	}
	return b.String()
}

func (x *lengthIndex) fake(s *drawstate.State, frames *frameStack, word []rune) string {
	key := fakeKey{x, strings.ToLower(string(word))}
	if fake, drawn := frames.fakes[key]; drawn {
		return fake
	}
	if frames.fakes == nil {
		frames.fakes = map[fakeKey]string{}
	}
	from, to := x.nearest(len(word))
	fake := x.drawIn(s, from, to)
	frames.fakes[key] = fake
	return fake
}

// nearest is the span of the texts whose length is n or, where none is, nearest n, the shorter on
// a tie.
func (x *lengthIndex) nearest(n int) (from, to int) {
	from, to = x.span(n, n)
	if from < to {
		return from, to
	}
	l := x.lengths[len(x.lengths)-1]
	switch {
	case from == len(x.lengths):
	case from == 0 || x.lengths[from]-n < n-x.lengths[from-1]:
		l = x.lengths[from]
	default:
		l = x.lengths[from-1]
	}
	return x.span(l, l)
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
