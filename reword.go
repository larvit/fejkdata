package fejkdata

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/drawstate"
)

// fakeKey is a word a {reword()} call met, in lower case, and where it read the words it drew a
// fake from.
type fakeKey struct {
	at   wordsAt
	word string
}

// wordsAt is where a {reword()} call reads its words: the node its path starts at, and the rest.
// A render keys its fakes by it, so calls reading one list keep one fake per word, whichever load
// built their words.
type wordsAt struct {
	start node
	tail  string
}

// wordList is the words a {reword()} call draws from, by length, and where it reads them.
type wordList struct {
	lengthIndex
	at wordsAt
}

// rewordDraw is a {reword()} call: the text it rewrites, its first operand, and the words its
// second operand's path holds, read at load.
type rewordDraw struct {
	text  *arm
	words *wordList
}

// rewordDrawOf builds the draw of o, a {reword()} call; built holds the words of each path read
// so far, so calls reading one path share its words.
func (t *template) rewordDrawOf(o *op, l *links, built map[wordsAt]*wordList) (*rewordDraw, error) {
	words, err := t.wordsOf(o.operands[1], l, built)
	if err != nil {
		return nil, err
	}
	return &rewordDraw{text: &o.operands[0], words: words}, nil
}

func (d *rewordDraw) draw(s *drawstate.State, t *template, env renderEnv, _ string) string {
	return d.words.reword(s, env.frames, readField(s, t, env, *d.text).text)
}

func (d *rewordDraw) mayFallBack() bool { return false }

// wordsOf is the values a's path renders that are made of letters alone, by length.
func (t *template) wordsOf(a arm, l *links, built map[wordsAt]*wordList) (*wordList, error) {
	if a.kind == namedRead {
		return nil, fmt.Errorf("%q is a name, which holds one pick; pass the path it binds, as in {%s(t,/en_US.word)}", a.head, builtinfunc.RewordName)
	}
	at := wordsAt{t.startOf(a.head, l), strings.Join(a.tail, ".")}
	if words, done := built[at]; done {
		return words, nil
	}
	w := valueWalk{links: l, seen: map[node]bool{}}
	if err := w.walk(at.start, a.tail, 1, "", ""); err != nil {
		return nil, fmt.Errorf("its words are read whole at load, and %w", err)
	}
	var letters []weightedText
	for _, v := range w.out {
		if v.text != "" && !strings.ContainsFunc(v.text, func(r rune) bool { return !unicode.IsLetter(r) }) {
			letters = append(letters, v)
		}
	}
	if len(letters) == 0 {
		return nil, fmt.Errorf("%s holds no word made of letters alone", a.spelling)
	}
	words := &wordList{newLengthIndex(letters), at}
	built[at] = words
	return words, nil
}

// reword rewrites each run of letters in text as a word of its length, or of the nearest length
// x holds, one fake per word in the render; every other rune stays.
func (x *wordList) reword(s *drawstate.State, frames *frameStack, text string) string {
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

func (x *wordList) fake(s *drawstate.State, frames *frameStack, word []rune) string {
	key := fakeKey{x.at, strings.ToLower(string(word))}
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
