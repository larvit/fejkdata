package grammar

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// The words that open a token's clauses, in the order they come: {path where … distinct N … as name}.
const (
	whereWord    = "where"
	distinctWord = "distinct"
)

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

// Distinct is a draw of N distinct values in one pick, printed joined by Sep where Joins is set.
type Distinct struct {
	N     int
	Joins bool
	Sep   string
}

var whereKeywords = []string{"and", "containing", "length", "not", "to"}

// containsWord reports whether s holds word after a space, where the word ends s or a space follows it.
func containsWord(s, word string) bool {
	return strings.Contains(s, " "+word+" ") || strings.HasSuffix(s, " "+word)
}

// cutWord cuts a token's body around its first " "+word outside a selector and a quote, where the
// word ends the body or a space follows it. A quote opens only inside a clause, so a path may hold an
// apostrophe, as o'neil does; inClause says body starts inside one.
func cutWord(body, word string, inClause bool) (head, rest string, found bool) {
	depth, quoted := 0, false
	for i := 0; i < len(body); i++ {
		switch c := body[i]; {
		case quoted:
			quoted = c != '\''
		case c == '[':
			depth++
		case c == ']' && depth > 0:
			depth--
		case depth > 0:
		case c == '\'' && inClause:
			quoted = true
		case atWord(body, i, word):
			return body[:i], strings.TrimPrefix(body[i+1+len(word):], " "), true
		case atWord(body, i, whereWord) || atWord(body, i, distinctWord):
			inClause = true
		}
	}
	return body, "", false
}

// atWord reports whether body holds " "+word at i, ending body or followed by a space, and not right
// after " as", where word is the name bound, as in {x as where}.
func atWord(body string, i int, word string) bool {
	after, found := strings.CutPrefix(body[i:], " "+word)
	return found && (after == "" || after[0] == ' ') && !strings.HasSuffix(body[:i+1], AsWord)
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

// clauses is a token body cut around its where and distinct clauses and the name it binds.
type clauses struct {
	head, where, distinct, name string
	filters, picks, binds       bool
}

// cutClauses cuts body around its clauses; found is false for a body with neither a where nor a
// distinct clause.
func cutClauses(body string) (c clauses, found bool) {
	c.head, c.where, c.filters = cutWord(body, whereWord, false)
	if c.filters {
		c.where, c.distinct, c.picks = cutWord(c.where, distinctWord, true)
	} else {
		c.head, c.distinct, c.picks = cutWord(body, distinctWord, false)
	}
	last := &c.where
	if c.picks {
		last = &c.distinct
	}
	*last, c.name, c.binds = lastAsOutsideQuotes(*last)
	return c, c.filters || c.picks
}

// clauseToken parses a token body carrying a where or a distinct clause: one path, its clauses, and
// the name it may bind.
func clauseToken(body string, c clauses) (Token, error) {
	_, _, bindsInHead := cutOutside(c.head, AsWord)
	switch {
	case indexOutside(c.head, '(') >= 0:
		return Token{}, fmt.Errorf("token {%s}: a where or distinct clause draws from a path, not a call", body)
	case indexOutside(c.head, '|') >= 0:
		return Token{}, fmt.Errorf("token {%s}: a where or distinct clause draws from one path; write a token for each", body)
	case bindsInHead:
		return Token{}, fmt.Errorf("token {%s}: as comes after the clauses: {path where … distinct N as name}", body)
	case c.filters && containsWord(c.head, distinctWord):
		return Token{}, fmt.Errorf("token {%s}: the where clause comes before distinct: {path where … distinct N}", body)
	}
	tok := Token{Kind: PathRead, Body: body, Arms: []string{c.head}}
	if c.binds {
		tok = Token{Kind: NameBind, Body: body, BoundRef: c.head, Bound: c.name}
	}
	var err error
	if c.filters {
		if tok.Where, err = ParseWhere(c.where); err != nil {
			return Token{}, fmt.Errorf("token {%s}: %w", body, err)
		}
	}
	if c.picks {
		if tok.Distinct, err = parseDistinct(c.distinct); err != nil {
			return Token{}, fmt.Errorf("token {%s}: %w", body, err)
		}
		if !c.binds && !tok.Distinct.Joins {
			n := tok.Distinct.N
			return Token{}, fmt.Errorf("token {%s}: distinct %d draws %d values into one token; join them, as distinct %d joined ',', or bind them, as distinct %d as name", body, n, n, n, n)
		}
	}
	return tok, nil
}

// parseDistinct parses a distinct clause: a whole number of 1 or more, then joined and a quoted
// separator where the values print as one.
func parseDistinct(clause string) (*Distinct, error) {
	lx, err := lexWhere(clause)
	if err != nil {
		return nil, fmt.Errorf("distinct %s: %w", clause, err)
	}
	p := &whereParser{lx: lx}
	l, ok := p.next()
	n, convErr := strconv.Atoi(l.text)
	if !ok || l.quoted || convErr != nil || n < 1 {
		return nil, fmt.Errorf("distinct %s: distinct takes a whole number of 1 or more, as distinct 3", clause)
	}
	d := &Distinct{N: n}
	if p.word("joined") {
		sep, ok := p.next()
		if !ok || !sep.quoted {
			return nil, fmt.Errorf("distinct %s: joined takes a quoted separator, as joined ','", clause)
		}
		d.Joins, d.Sep = true, sep.text
	}
	if !p.done() {
		return nil, fmt.Errorf("distinct %s: %s follows the count, where only joined ',' may", clause, p.peek())
	}
	return d, nil
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
				return nil, fmt.Errorf("a quote opens at %q and never closes; a quoted value can't hold an apostrophe", clause[i:])
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
	if err := CheckIdentifier(withoutMember(l.text)); err != nil {
		return Length{}, fmt.Errorf("length %s: %w", l.text, err)
	}
	return Length{Name: l.text}, nil
}

// withoutMember is name without a member's key, as m of m[1]; the key is checked where the name
// resolves.
func withoutMember(name string) string {
	if head, _, found := strings.Cut(name, "["); found && strings.HasSuffix(name, "]") {
		return head
	}
	return name
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
