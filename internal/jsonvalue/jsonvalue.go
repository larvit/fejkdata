// Package jsonvalue holds a decoded JSON value whose numbers keep the text they are written in.
package jsonvalue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"unicode/utf8"
)

// Value is a JSON value that only Decode, String and SortedObject make, so a template never
// compiles from a value encoding/json decoded with its numbers as float64. The zero Value is JSON null.
type Value struct{ v any }

// Object is a JSON object: its fields by key, and its keys in the order written. A key written
// twice keeps its first place and its last value.
type Object struct {
	Keys   []string
	Fields map[string]any
}

// Decode decodes one JSON value, keeping each number as the json.Number it is written as and each
// object's keys in the order written.
func Decode(b []byte) (Value, error) {
	if !json.Valid(b) {
		var v any
		return Value{}, located(b, json.Unmarshal(b, &v))
	}
	p := parser{b: b}
	return Value{p.value()}, nil
}

// parser reads JSON that json.Valid passed, so it checks no syntax. It allocates less than
// json.Decoder's tokens do, which New's allocation gate holds it to.
type parser struct {
	b []byte
	i int
}

func (p *parser) space() {
	for p.i < len(p.b) && bytes.IndexByte([]byte(" \t\r\n"), p.b[p.i]) >= 0 {
		p.i++
	}
}

// value reads the next value, after any whitespace.
func (p *parser) value() any {
	p.space()
	switch p.b[p.i] {
	case '{':
		return p.object()
	case '[':
		return p.list()
	case '"':
		return p.str()
	case 't':
		p.i += len("true")
		return true
	case 'f':
		p.i += len("false")
		return false
	case 'n':
		p.i += len("null")
		return nil
	}
	return p.number()
}

func (p *parser) object() Object {
	o := Object{Fields: map[string]any{}}
	for p.i++; !p.closes('}'); {
		key := p.str()
		p.space()
		p.i++
		if _, twice := o.Fields[key]; !twice {
			o.Keys = append(o.Keys, key)
		}
		o.Fields[key] = p.value()
	}
	return o
}

func (p *parser) list() []any {
	list := []any{}
	for p.i++; !p.closes(']'); {
		list = append(list, p.value())
	}
	return list
}

// closes steps past whitespace and a comma to the next member, or past end where the object or
// list ends there.
func (p *parser) closes(end byte) bool {
	p.space()
	if p.b[p.i] == ',' {
		p.i++
		p.space()
	}
	if p.b[p.i] != end {
		return false
	}
	p.i++
	return true
}

func (p *parser) str() string {
	start, escaped := p.i, false
	for p.i++; p.b[p.i] != '"'; p.i++ {
		if p.b[p.i] == '\\' {
			escaped = true
			p.i++
		}
	}
	p.i++
	if raw := p.b[start+1 : p.i-1]; !escaped && utf8.Valid(raw) {
		return string(raw)
	}
	var s string
	_ = json.Unmarshal(p.b[start:p.i], &s)
	return s
}

func (p *parser) number() json.Number {
	start := p.i
	for p.i < len(p.b) && bytes.IndexByte([]byte("+-.eE0123456789"), p.b[p.i]) >= 0 {
		p.i++
	}
	return json.Number(p.b[start:p.i])
}

// located prefixes a syntax error with the line and column where decoding stopped.
func located(b []byte, err error) error {
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) {
		return err
	}
	before := b[:max(syntax.Offset-1, 0)]
	line := bytes.Count(before, []byte("\n")) + 1
	column := utf8.RuneCount(before[bytes.LastIndexByte(before, '\n')+1:]) + 1
	return fmt.Errorf("line %d, column %d: %w", line, column, err)
}

// String is the JSON string s.
func String(s string) Value { return Value{s} }

// SortedObject is the JSON object holding fields, its keys in name order.
func SortedObject(fields map[string]Value) Value {
	o := Object{Fields: make(map[string]any, len(fields))}
	for k, f := range fields {
		o.Keys = append(o.Keys, k)
		o.Fields[k] = f.v
	}
	slices.Sort(o.Keys)
	return Value{o}
}

// Any is the value as Decode gives it: an Object, []any, string, json.Number, bool or nil.
func (v Value) Any() any { return v.v }

// Fields is the fields of v where v is an Object.
func Fields(v any) (fields map[string]any, isObject bool) {
	o, isObject := v.(Object)
	return o.Fields, isObject
}

// Kind names the kind of a value as Any gives it, in the data format's own terms, such as
// "a number" or "a list".
func Kind(v any) string {
	switch v.(type) {
	case json.Number:
		return "a number"
	case bool:
		return "a boolean"
	case string:
		return "a string"
	case []any:
		return "a list"
	case Object:
		return "an object"
	case nil:
		return "null"
	}
	return fmt.Sprintf("%T", v)
}
