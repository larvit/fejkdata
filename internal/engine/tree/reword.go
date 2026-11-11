package tree

// WordsAt is where a {reword()} call reads its words: the node its path starts at, and the rest.
// A render keys its fakes by it, so calls reading one list keep one fake per word, whichever load
// built their words.
type WordsAt struct {
	start Node
	tail  string
}

func NewWordsAt(start Node, tail string) WordsAt { return WordsAt{start, tail} }

func (w WordsAt) Start() Node { return w.start }

// WordList is the words a {reword()} call draws from, by length, and where it reads them.
type WordList struct {
	index LengthIndex
	at    WordsAt
}

func NewWordList(index LengthIndex, at WordsAt) *WordList { return &WordList{index, at} }

func (w *WordList) At() WordsAt { return w.at }

func (w *WordList) Index() *LengthIndex { return &w.index }

// RewordDraw is a {reword()} call: the text it rewrites, its first operand, and the words its
// second operand's path holds, read at load.
type RewordDraw struct {
	text  *Arm
	words *WordList
}

func NewRewordDraw(text *Arm, words *WordList) *RewordDraw { return &RewordDraw{text, words} }

func (*RewordDraw) isTokenDraw() {}

func (d *RewordDraw) MayFallBack() bool { return false }

func (d *RewordDraw) Text() *Arm { return d.text }

func (d *RewordDraw) Words() *WordList { return d.words }
