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
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	return Value{decodeNext(dec)}, nil
}

// decodeNext reads the next value of JSON that json.Valid passed, so no token fails.
func decodeNext(dec *json.Decoder) any {
	switch tok, _ := dec.Token(); tok {
	case json.Delim('{'):
		return decodeObject(dec)
	case json.Delim('['):
		list := []any{}
		for dec.More() {
			list = append(list, decodeNext(dec))
		}
		_, _ = dec.Token()
		return list
	default:
		return tok
	}
}

func decodeObject(dec *json.Decoder) Object {
	o := Object{Fields: map[string]any{}}
	for dec.More() {
		tok, _ := dec.Token()
		key := tok.(string)
		if _, twice := o.Fields[key]; !twice {
			o.Keys = append(o.Keys, key)
		}
		o.Fields[key] = decodeNext(dec)
	}
	_, _ = dec.Token()
	return o
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
