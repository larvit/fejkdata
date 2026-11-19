package builtinfunc

import (
	"slices"
	"testing"
)

func TestCheckArgsReturnsTheNumbersItParsed(t *testing.T) {
	for _, c := range []struct {
		name   string
		args   []string
		ints   []int
		floats []float64
	}{
		{"digits", []string{"4"}, []int{4}, nil},
		{"hex", []string{"0"}, []int{0}, nil},
		{"luhn", []string{"6"}, []int{6}, nil},
		{"luhn", nil, nil, nil},
		{"int", []string{"-3", "7"}, []int{-3, 7}, nil},
		{"float", []string{"1.5", "2.5", "2"}, []int{2}, []float64{1.5, 2.5}},
		{"born", []string{"18", "65", "'2006'"}, []int{18, 65}, nil},
		{"tick", []string{"2026-01-01", "60", "10", "'2006'"}, []int{60, 10}, nil},
		{"calc", []string{"a"}, nil, nil},
		{"calc", []string{"a", "2"}, []int{2}, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			p, err := builtins[c.name].checkArgs(c.args)
			if err != nil {
				t.Fatalf("checkArgs(%s, %q) = %v, want the args accepted", c.name, c.args, err)
			}
			if !slices.Equal(p.ints, c.ints) || !slices.Equal(p.floats, c.floats) {
				t.Errorf("checkArgs(%s, %q) parsed ints %v floats %v, want ints %v floats %v", c.name, c.args, p.ints, p.floats, c.ints, c.floats)
			}
		})
	}
}
