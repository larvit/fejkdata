// Package grammar is how the template language is written: format tokens, paths and
// their selectors, reference sigils, identifiers, calc expressions, and whether an argument is a
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
		var u scanUnit
		switch c := rs[i]; c {
		case '{':
			if i+1 < len(rs) && rs[i+1] == '{' {
				u.char = '{'
				i++
				break
			}
			end := i + 1
			for end < len(rs) && rs[end] != '}' {
				if rs[end] == '{' {
					return fmt.Errorf("'{' inside a token in %q; a literal brace is written {{", format)
				}
				end++
			}
			if end >= len(rs) {
				return fmt.Errorf("unterminated '{' in %q", format)
			}
			u.isToken, u.body = true, string(rs[i+1:end])
			i = end
		case '}':
			if i+1 < len(rs) && rs[i+1] == '}' {
				u.char = '}'
				i++
				break
			}
			return fmt.Errorf("lone '}' in %q; a literal brace is written }}", format)
		default:
			u.char = c
		}
		if err := fn(u); err != nil {
			return err
		}
	}
	return nil
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
		if ref, name, binds := cutOutside(u.body, AsWord); binds && indexOutside(u.body, '(') < 0 {
			toks = append(toks, Token{Kind: NameBind, Body: u.body, BoundRef: ref, Bound: name})
			return nil
		}
		if indexOutside(u.body, '(') < 0 {
			toks = append(toks, Token{Kind: PathRead, Body: u.body, Arms: splitOutside(u.body, '|')})
			return nil
		}
		name, args, ok := FuncCall(u.body)
		if !ok {
			return fmt.Errorf("malformed function token {%s}", u.body)
		}
		toks = append(toks, Token{Kind: BuiltinCall, Body: u.body, Fn: name, Args: args})
		return nil
	})
	if err != nil {
		return nil, err
	}
	flush()
	return toks, nil
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
	depth, quoted, start := 0, false, 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\'' && depth == 0:
			quoted = !quoted
		case quoted:
		case c == '[':
			depth++
		case c == ']' && depth > 0:
			depth--
		case c == ',' && depth == 0:
			args, start = append(args, strings.TrimSpace(s[start:i])), i+1
		}
	}
	return append(args, strings.TrimSpace(s[start:]))
}
