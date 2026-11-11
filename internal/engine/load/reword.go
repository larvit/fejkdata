package load

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// rewordDrawOf builds the draw of o, a {reword()} call; built holds the words of each path read
// so far, so calls reading one path share its words.
func rewordDrawOf(t *tree.Template, o *tree.Op, l *tree.LoadFinds, built map[tree.WordsAt]*tree.WordList) (*tree.RewordDraw, error) {
	words, err := wordsOf(t, o.Operands()[1], l, built)
	if err != nil {
		return nil, err
	}
	return tree.NewRewordDraw(&o.Operands()[0], words), nil
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
