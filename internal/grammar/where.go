package grammar

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// WhereWord opens a token's filter: {path where clause}, or {path where clause as name}.
const WhereWord = " where "

// Where is a filter over the values a path can draw: a value passes when every condition holds.
type Where struct {
	Length        *LengthRange // nil where the clause names no length
	Not           []string     // values refused whole
	NotContaining []string     // values refused wherever they hold one of these
}

// LengthRange is the lengths a value may have, in code points, both ends included.
type LengthRange struct{ Lo, Hi Length }

// Length is a whole number, or the name it reads at render where Name is set.
type Length struct {
	N    int
	Name string
}

var whereKeywords = []string{"and", "containing", "length", "not", "to"}

// containsWhere reports whether s holds the word that opens a filter.
func containsWhere(s string) bool {
	return strings.Contains(s, WhereWord) || strings.HasSuffix(s, strings.TrimSuffix(WhereWord, " "))
}

// cutWhere cuts a token's body around its first " where" outside a selector and a quote, where
// the word ends the body or a space follows it.
func cutWhere(body string) (head, clause string, found bool) {
	word := strings.TrimSuffix(WhereWord, " ")
	depth, quoted := 0, false
	for i := 0; i < len(body); i++ {
		switch c := body[i]; {
		case c == '[' && !quoted:
			depth++
		case c == ']' && depth > 0:
			depth--
		case c == '\'' && depth == 0:
			quoted = !quoted
		case depth == 0 && !quoted && strings.HasPrefix(body[i:], word):
			rest := body[i+len(word):]
			if rest == "" || rest[0] == ' ' {
				return body[:i], strings.TrimPrefix(rest, " "), true
			}
		}
	}
	return body, "", false
}

// lastAsOutsideQuotes cuts clause around its last " as " outside a quote.
func lastAsOutsideQuotes(clause string) (before, name string, found bool) {
	at, quoted := -1, false
	for i := 0; i < len(clause); i++ {
		switch {
		case clause[i] == '\'':
			quoted = !quoted
		case !quoted && strings.HasPrefix(clause[i:], AsWord):
			at = i
		}
	}
	if at < 0 {
		return clause, "", false
	}
	return clause[:at], clause[at+len(AsWord):], true
}

// whereToken parses a token body carrying a filter: one path, a clause, and the name it may bind.
func whereToken(body, head, clause string) (Token, error) {
	switch {
	case indexOutside(head, '(') >= 0:
		return Token{}, fmt.Errorf("token {%s}: a where clause filters a path's draw, not a call's", body)
	case indexOutside(head, '|') >= 0:
		return Token{}, fmt.Errorf("token {%s}: a where clause filters one path; write a token for each", body)
	}
	if _, _, binds := cutOutside(head, AsWord); binds {
		return Token{}, fmt.Errorf("token {%s}: the where clause comes before as: {path where … as name}", body)
	}
	clause, name, binds := lastAsOutsideQuotes(clause)
	w, err := ParseWhere(clause)
	if err != nil {
		return Token{}, fmt.Errorf("token {%s}: %w", body, err)
	}
	if binds {
		return Token{Kind: NameBind, Body: body, BoundRef: head, Bound: name, Where: w}, nil
	}
	return Token{Kind: PathRead, Body: body, Arms: []string{head}, Where: w}, nil
}

type whereLexeme struct {
	text   string
	quoted bool
}

func lexWhere(clause string) ([]whereLexeme, error) {
	var out []whereLexeme
	for i := 0; i < len(clause); {
		switch c := clause[i]; c {
		case ' ':
			i++
		case ',':
			out = append(out, whereLexeme{text: ","})
			i++
		case '\'':
			end := strings.IndexByte(clause[i+1:], '\'')
			if end < 0 {
				return nil, fmt.Errorf("a quote opens at %q and never closes", clause[i:])
			}
			out = append(out, whereLexeme{text: clause[i+1 : i+1+end], quoted: true})
			i += end + 2
		default:
			end := strings.IndexAny(clause[i:], " ,'")
			if end < 0 {
				end = len(clause) - i
			}
			out = append(out, whereLexeme{text: clause[i : i+end]})
			i += end
		}
	}
	return out, nil
}

type whereParser struct {
	lx []whereLexeme
	at int
}

func (p *whereParser) done() bool { return p.at == len(p.lx) }

func (p *whereParser) next() (whereLexeme, bool) {
	if p.done() {
		return whereLexeme{}, false
	}
	p.at++
	return p.lx[p.at-1], true
}

func (p *whereParser) peek() string {
	if p.done() {
		return "the end"
	}
	return strconv.Quote(p.lx[p.at].text)
}

// word consumes the bare word w where it comes next.
func (p *whereParser) word(w string) bool {
	if p.done() || p.lx[p.at].quoted || p.lx[p.at].text != w {
		return false
	}
	p.at++
	return true
}

// ParseWhere parses a where clause: conditions joined by and, each `length N`, `length N to M`,
// `not 'a', 'b'` or `not containing 'a', 'b'`, where a length is a whole number or a name.
func ParseWhere(clause string) (*Where, error) {
	lx, err := lexWhere(clause)
	if err != nil {
		return nil, fmt.Errorf("where %s: %w", clause, err)
	}
	p := &whereParser{lx: lx}
	w := &Where{}
	for {
		if err := p.condition(w); err != nil {
			return nil, fmt.Errorf("where %s: %w", clause, err)
		}
		if p.done() {
			return w, nil
		}
		if !p.word("and") {
			return nil, fmt.Errorf("where %s: and joins two conditions; %s follows one", clause, p.peek())
		}
	}
}

func (p *whereParser) condition(w *Where) error {
	switch {
	case p.word("length"):
		if w.Length != nil {
			return errors.New("a where clause takes one length")
		}
		r, err := p.lengthRange()
		w.Length = r
		return err
	case p.word("not"):
		into := &w.Not
		if p.word("containing") {
			into = &w.NotContaining
		}
		values, err := p.quotedList()
		*into = append(*into, values...)
		return err
	}
	return fmt.Errorf("a condition starts with length or not, not %s", p.peek())
}

func (p *whereParser) lengthRange() (*LengthRange, error) {
	lo, err := p.length()
	if err != nil {
		return nil, err
	}
	hi := lo
	if p.word("to") {
		if hi, err = p.length(); err != nil {
			return nil, err
		}
	}
	if lo.Name == "" && hi.Name == "" && lo.N > hi.N {
		return nil, fmt.Errorf("length %d to %d holds no length; write the shorter end first", lo.N, hi.N)
	}
	return &LengthRange{Lo: lo, Hi: hi}, nil
}

func (p *whereParser) length() (Length, error) {
	l, ok := p.next()
	if !ok || l.quoted || l.text == "," || slices.Contains(whereKeywords, l.text) {
		return Length{}, fmt.Errorf("length takes a whole number or a name, as length 5 or length k")
	}
	if l.text[0] == '-' || l.text[0] >= '0' && l.text[0] <= '9' {
		n, err := strconv.Atoi(l.text)
		if err != nil || n < 0 {
			return Length{}, fmt.Errorf("length %s is no whole number of 0 or more", l.text)
		}
		return Length{N: n}, nil
	}
	if err := CheckIdentifier(l.text); err != nil {
		return Length{}, fmt.Errorf("length %s: %w", l.text, err)
	}
	return Length{Name: l.text}, nil
}

func (p *whereParser) quotedList() ([]string, error) {
	var values []string
	for {
		l, ok := p.next()
		if !ok || !l.quoted {
			return nil, errors.New("not takes quoted values, as not 'cat', 'dog'")
		}
		values = append(values, l.text)
		if p.done() || p.lx[p.at].quoted || p.lx[p.at].text != "," {
			return values, nil
		}
		p.at++
	}
}
