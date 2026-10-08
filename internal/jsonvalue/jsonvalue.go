// Package jsonvalue holds a decoded JSON value whose numbers keep the text they are written in.
package jsonvalue

import (
	"bytes"
	"encoding/json"
)

// Value is a JSON value that only Decode, String and Object make, so a template never compiles
// from a value encoding/json decoded with its numbers as float64. The zero Value is JSON null.
type Value struct{ v any }

// Decode decodes one JSON value, keeping each number as the json.Number it is written as.
func Decode(b []byte) (Value, error) {
	var v any
	if !json.Valid(b) {
		return Value{}, json.Unmarshal(b, &v)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	err := dec.Decode(&v)
	return Value{v}, err
}

// String is the JSON string s.
func String(s string) Value { return Value{s} }

// Object is the JSON object holding fields.
func Object(fields map[string]Value) Value {
	m := make(map[string]any, len(fields))
	for k, f := range fields {
		m[k] = f.v
	}
	return Value{m}
}

// Any is the value as encoding/json decodes it: a map[string]any, []any, string, json.Number,
// bool or nil.
func (v Value) Any() any { return v.v }
