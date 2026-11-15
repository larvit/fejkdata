// Package grammar is how the template language is written: format tokens, paths and
// their selectors, reference sigils, identifiers, calc expressions, and whether a string is a
// template or a path. It reads strings only.
package grammar

import (
	"fmt"
	"strings"
)

// scanUnit is one unit of a scanned format string: a literal rune or the body of a
// {…} token.
type scanUnit struct {
	isToken bool // a {…} body, else a literal rune
	char    rune
	body    string
}

// eachScanUnit scans a format string once and calls fn for each unit, the single
// source of truth for how braces are read: "{{" and "}}" are literal braces, a "{"
// opens a token that must reach its "}", and a lone "}" is an error.
func eachScanUnit(format string, fn func(scanUnit) error) error {
	rs := []rune(format)
	for i := 0; i < len(rs); i++ {
		u, last, err := scanOne(rs, i, format)
		if err != nil {
			return err
		}
		i = last
		if err := fn(u); err != nil {
			return err
		}
	}
	return nil
}

// scanOne scans the unit of rs that starts at index i, and returns it with the index of its last rune.
func scanOne(rs []rune, i int, format string) (u scanUnit, last int, err error) {
	c := rs[i]
	switch c {
	case '{':
		if doubled(rs, i) {
			return scanUnit{char: c}, i + 1, nil
		}
		end, err := tokenEnd(rs, i, format)
		if err != nil {
			return scanUnit{}, 0, err
		}
		return scanUnit{isToken: true, body: string(rs[i+1 : end])}, end, nil
	case '}':
		if doubled(rs, i) {
			return scanUnit{char: c}, i + 1, nil
		}
		return scanUnit{}, 0, fmt.Errorf("lone '}' in %q; a literal brace is written }}", format)
	}
	return scanUnit{char: c}, i, nil
}

// doubled reports whether the rune after index i repeats the one at i.
func doubled(rs []rune, i int) bool { return i+1 < len(rs) && rs[i+1] == rs[i] }

// tokenEnd returns the index of the "}" closing the token that the "{" at open in format starts.
func tokenEnd(rs []rune, open int, format string) (int, error) {
	for end := open + 1; end < len(rs); end++ {
		switch rs[end] {
		case '}':
			return end, nil
		case '{':
			return 0, fmt.Errorf("'{' inside a token in %q; a literal brace is written {{", format)
		}
	}
	return 0, fmt.Errorf("unterminated '{' in %q", format)
}

// Token is one parsed unit of a format.
type Token struct {
	Kind TokenKind
	Lit  string   // LiteralRun
	Body string   // the braces' content, as written
	Fn   string   // BuiltinCall
	Args []string // BuiltinCall
	Arms []string // PathRead: the '|' alternatives
	// NameBind: the reference it binds, and the name.
	BoundRef, Bound string
	Where           *Where    // PathRead, NameBind or BuiltinCall: the filter on its draw, nil for none
	Distinct        *Distinct // PathRead or NameBind: several distinct draws in one pick, nil for one draw
}

type TokenKind uint8

const (
	BuiltinCall TokenKind = iota + 1
	LiteralRun
	NameBind
	PathRead
)

const AsWord = " as "

// ParseFormat is the one reading of a format's tokens: a '(' outside a selector makes
// a token a call.
func ParseFormat(format string) ([]Token, error) {
	var toks []Token
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			toks = append(toks, Token{Kind: LiteralRun, Lit: lit.String()})
			lit.Reset()
		}
	}
	err := eachScanUnit(format, func(u scanUnit) error {
		if !u.isToken {
			lit.WriteRune(u.char)
			return nil
		}
		flush()
		tok, err := bodyToken(u.body)
		toks = append(toks, tok)
		return err
	})
	if err != nil {
		return nil, err
	}
	flush()
	return toks, nil
}

// bodyToken parses the body of a {…} token: a path or name read with its clauses, a name bound, a
// path read, or a call.
func bodyToken(body string) (Token, error) {
	if c, found := cutClauses(body); found {
		return clauseToken(body, c)
	}
	isCall := indexOutside(body, '(') >= 0
	if ref, name, binds := cutOutside(body, AsWord); binds && !isCall {
		return Token{Kind: NameBind, Body: body, BoundRef: ref, Bound: name}, nil
	}
	if !isCall {
		return Token{Kind: PathRead, Body: body, Arms: splitOutside(body, '|')}, nil
	}
	name, args, ok := FuncCall(body)
	if !ok {
		return Token{}, fmt.Errorf("malformed function token {%s}", body)
	}
	return Token{Kind: BuiltinCall, Body: body, Fn: name, Args: args}, nil
}

// FuncCall splits a "{token}" body shaped name(args) into its parts; ok is false
// for a PathRead body. A '(' without a trailing ')' yields ok=false.
func FuncCall(body string) (name string, args []string, ok bool) {
	lp := indexOutside(body, '(')
	if lp < 0 || !strings.HasSuffix(body, ")") {
		return "", nil, false
	}
	return body[:lp], splitArgs(body[lp+1 : len(body)-1]), true
}

// splitArgs parses a function arg list: comma-separated outside a selector or a
// quoted layout, trimmed; empty -> none.
func splitArgs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var args []string
	start := 0
	(&scanner{quotes: true}).find(s, func(i int) bool {
		if s[i] == ',' {
			args, start = append(args, strings.TrimSpace(s[start:i])), i+1
		}
		return false
	})
	return append(args, strings.TrimSpace(s[start:]))
}
