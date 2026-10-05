package grammar

import (
	"fmt"
	"strings"
)

// reservedInName is what a category, folder or field name may not contain: a dot
// separates the segments of a path, '|' the arms of a token, '(' opens a function
// call, braces delimit the token, '/' starts a reference, and brackets and a quote
// open a JSON value. A name carrying one is rejected where it is authored rather
// than where it would be unreachable.
const reservedInName = ".|({}/[]\""

var reservedList = strings.Join(strings.Split(reservedInName, ""), " ")

// CheckName rejects a name the dot path, {token} and JSON grammars cannot spell, or a struct
// tag cannot read.
// Both a category or folder and a field go through it, so there is one answer to
// what a name may contain.
func CheckName(name string) error {
	if name == "" {
		return fmt.Errorf("%q is empty, which is not a path segment, so List never offers it", name)
	}
	if name == "-" {
		return fmt.Errorf(`%q is reserved: the struct tag fake:"-" leaves a field unfilled, so no tag could read it; rename it`, name)
	}
	if strings.Contains(name, AsWord) {
		return fmt.Errorf("%q contains %q, which a token reads as binding a name; rename it", name, AsWord)
	}
	if i := strings.IndexAny(name, reservedInName); i >= 0 {
		return fmt.Errorf("%q contains %q; a name may not use %s, which the dot path, {token} and JSON grammars reserve",
			name, name[i:i+1], reservedList)
	}
	return nil
}

// CheckPathNames rejects a dotted path with a segment no name may be.
func CheckPathNames(path string) error {
	segs, err := SplitPath(path)
	if err != nil {
		return err
	}
	for _, seg := range NameSegments(segs) {
		if err := CheckName(seg); err != nil {
			return fmt.Errorf("path %w", err)
		}
	}
	return nil
}
