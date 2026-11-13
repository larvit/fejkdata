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

// Where is a filter over the values a path or a call can draw: a value passes when every condition
// holds. A name's string is read at render, and Else is printed where no value passes.
type Where struct {
	Length             *LengthRange // nil where the clause names no length
	Not                []string     // quoted values refused whole
	NotNames           []string     // names whose strings are refused whole
	NotContaining      []string     // quoted substrings: a value holding one is refused
	NotContainingNames []string     // names whose strings a value may not hold
	Else               *Fallback    // nil where the clause names no fallback
}

// Fallback is the quoted Text a where clause prints where no value passes, or the name it reads where
// Name is set.
type Fallback struct{ Text, Name string }

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

var whereKeywords = []string{"and", "containing", "else", "length", "not", "to"}

// containsWord reports whether s holds word after a space, where the word ends s or a space follows it.
func containsWord(s, word string) bool {
	return strings.Contains(s, " "+word+" ") || strings.HasSuffix(s, " "+word)
}

// cutWord cuts a token's body around its first " "+word outside a selector and a quote, where the
// word ends the body or a space follows it. A quote opens only inside a clause; inClause says body
// starts inside one.
func cutWord(body, word string, inClause bool) (head, rest string, found bool) {
	sc := scanner{quotes: inClause}
	i := sc.find(body, func(i int) bool {
		if atWord(body, i, word) {
			return true
		}
		sc.quotes = sc.quotes || atWord(body, i, whereWord) || atWord(body, i, distinctWord)
		return false
	})
	if i < 0 {
		return body, "", false
	}
	return body[:i], strings.TrimPrefix(body[i+1+len(word):], " "), true
}

// atWord reports whether body holds " "+word at i, ending body or followed by a space, and not right
// after " as", where word is the name bound, as in {x as where}.
func atWord(body string, i int, word string) bool {
	after, found := strings.CutPrefix(body[i:], " "+word)
	return found && (after == "" || after[0] == ' ') && !strings.HasSuffix(body[:i+1], AsWord)
}

// lastAsOutside cuts clause around its last " as " outside a quote and a selector.
func lastAsOutside(clause string) (before, name string, found bool) {
	at := -1
	(&scanner{quotes: true}).find(clause, func(i int) bool {
		if strings.HasPrefix(clause[i:], AsWord) {
			at = i
		}
		return false
	})
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
	*last, c.name, c.binds = lastAsOutside(*last)
	return c, c.filters || c.picks
}

// clauseToken parses a token body carrying a where or a distinct clause: one path, its clauses, and
// the name it may bind.
func clauseToken(body string, c clauses) (Token, error) {
	if err := refuseClauseHead(body, c); err != nil {
		return Token{}, err
	}
	if indexOutside(c.head, '(') >= 0 {
		return callClauseToken(body, c)
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
		if tok.Distinct, err = distinctClause(body, c); err != nil {
			return Token{}, err
		}
	}
	return tok, nil
}

// distinctClause parses the distinct clause of the token body, refusing one that neither joins nor
// binds its values.
func distinctClause(body string, c clauses) (*Distinct, error) {
	d, err := parseDistinct(c.distinct)
	if err != nil {
		return nil, fmt.Errorf("token {%s}: %w", body, err)
	}
	if !c.binds && !d.Joins {
		n := d.N
		return nil, fmt.Errorf("token {%s}: distinct %d draws %d values into one token; join them, such as {w distinct %d joined ','}, or bind them, such as {w distinct %d as name}", body, n, n, n, n)
	}
	return d, nil
}

// refuseClauseHead refuses a head of the token body that draws from several paths, binds a name
// before its clauses, or puts distinct before where.
func refuseClauseHead(body string, c clauses) error {
	_, _, bindsInHead := cutOutside(c.head, AsWord)
	switch {
	case indexOutside(c.head, '|') >= 0:
		return fmt.Errorf("token {%s}: a where or distinct clause draws from one path; write a token for each", body)
	case bindsInHead:
		return fmt.Errorf("token {%s}: \"as\" comes after the clauses: {path where … distinct N as name}", body)
	case c.filters && containsWord(c.head, distinctWord):
		return fmt.Errorf("token {%s}: the where clause comes before distinct: {path where … distinct N}", body)
	}
	return nil
}

// callClauseToken parses a call carrying a where clause, which neither draws distinct values nor binds a
// name.
func callClauseToken(body string, c clauses) (Token, error) {
	fn, args, ok := FuncCall(c.head)
	switch {
	case !ok:
		return Token{}, fmt.Errorf("malformed function token {%s}", body)
	case c.picks:
		return Token{}, fmt.Errorf("token {%s}: a distinct clause draws from a path, not a call", body)
	case c.binds:
		return Token{}, fmt.Errorf("token {%s}: a call can't be bound; put the call and its clause in a field, and bind the field", body)
	}
	w, err := ParseWhere(c.where)
	if err != nil {
		return Token{}, fmt.Errorf("token {%s}: %w", body, err)
	}
	return Token{Kind: BuiltinCall, Body: body, Fn: fn, Args: args, Where: w}, nil
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
		return nil, fmt.Errorf("distinct %s: \"distinct\" takes a whole number of 1 or more, such as {w distinct 3 joined ','}", clause)
	}
	d := &Distinct{N: n}
	if p.word("joined") {
		sep, ok := p.next()
		if !ok || !sep.quoted {
			return nil, fmt.Errorf("distinct %s: \"joined\" takes a quoted separator, such as {w distinct 2 joined ','}", clause)
		}
		d.Joins, d.Sep = true, sep.text
	}
	if !p.done() {
		return nil, fmt.Errorf("distinct %s: %s follows the count, where only \"joined\" and a quoted separator may", clause, p.peek())
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
// `not 'a', n` or `not containing 'a', n`, where a length is a whole number or a name, and a list
// holds quoted strings and names. `else 'text'` or `else n` may end it.
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
		if p.word("else") {
			return w, p.fallback(w, clause)
		}
		if p.done() {
			return w, nil
		}
		if !p.word("and") {
			return nil, fmt.Errorf("where %s: \"and\" joins two conditions; %s follows one", clause, p.peek())
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
		if p.word("containing") {
			return p.stringList(&w.NotContaining, &w.NotContainingNames)
		}
		return p.stringList(&w.Not, &w.NotNames)
	}
	return fmt.Errorf("a condition starts with \"length\" or \"not\", not %s", p.peek())
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
		return Length{}, fmt.Errorf("\"length\" takes a whole number or a name, such as {w where length 5} or {w where length k}")
	}
	if startsNumber(l.text) {
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

// stringList appends the quoted strings of a list to quoted, and the names to names.
func (p *whereParser) stringList(quoted, names *[]string) error {
	for {
		l, ok := p.next()
		if err := listItem(l, ok, quoted, names); err != nil {
			return err
		}
		if p.done() || p.lx[p.at].quoted || p.lx[p.at].text != "," {
			return nil
		}
		p.at++
	}
}

// listItem appends the lexeme l of a not-list to quoted or names; ok is false past the clause's end.
func listItem(l whereLexeme, ok bool, quoted, names *[]string) error {
	switch {
	case ok && l.quoted:
		*quoted = append(*quoted, l.text)
	case ok && slices.Contains(whereKeywords, l.text):
		return fmt.Errorf("\"not\" takes quoted strings or names, such as {w where not 'cat', n}, and %q is a word of the clause, so a name read here needs another spelling", l.text)
	case !ok || l.text == ",":
		return errors.New("\"not\" takes quoted strings or names, such as {w where not 'cat', n}")
	case startsNumber(l.text):
		return fmt.Errorf("not %s: quote a string, as '%s'; a bare word is a name", l.text, l.text)
	default:
		if err := CheckIdentifier(withoutMember(l.text)); err != nil {
			return fmt.Errorf("not %s: %w", l.text, err)
		}
		*names = append(*names, l.text)
	}
	return nil
}

// fallback parses what follows else into w: a quoted text or a name, ending clause.
func (p *whereParser) fallback(w *Where, clause string) error {
	l, ok := p.next()
	f, err := fallbackOf(l, ok, clause)
	if err != nil {
		return err
	}
	w.Else = f
	if !p.done() {
		return fmt.Errorf("where %s: \"else\" and its fallback end the clause; %s follows them", clause, p.peek())
	}
	return nil
}

// fallbackOf reads the lexeme l after else of clause; ok is false past the clause's end.
func fallbackOf(l whereLexeme, ok bool, clause string) (*Fallback, error) {
	switch {
	case ok && !l.quoted && slices.Contains(whereKeywords, l.text):
		return nil, fmt.Errorf("where %s: \"else\" takes a quoted fallback or a name, such as {w where not 'cat' else 'none'}, and %q is a word of the clause, so a name read here needs another spelling", clause, l.text)
	case !ok || !l.quoted && (l.text == "," || startsNumber(l.text)):
		return nil, fmt.Errorf("where %s: \"else\" takes a quoted fallback or a name, such as {w where not 'cat' else 'none'}", clause)
	case l.quoted:
		return &Fallback{Text: l.text}, nil
	}
	if err := CheckIdentifier(withoutMember(l.text)); err != nil {
		return nil, fmt.Errorf("where %s: else %s: %w", clause, l.text, err)
	}
	return &Fallback{Name: l.text}, nil
}

func startsNumber(word string) bool { return word[0] == '-' || word[0] >= '0' && word[0] <= '9' }
