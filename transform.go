package fejkdata

import (
	"fmt"
	"strings"
	"unicode"
)

// transforms are the builtins that rewrite one operand's value; they nest, so
// {lowercase(ascii(x))} folds then lowers.
var transforms = map[string]func(string) string{
	"ascii":     asciiFold,
	"lowercase": strings.ToLower,
	"uppercase": strings.ToUpper,
}

// unwrapTransform peels nested transform calls off an operand arg, returning the
// field it finally names and the transforms to apply, innermost last.
func unwrapTransform(arg string) (leaf string, chain []func(string) string, err error) {
	for {
		name, args, isCall := funcCall(arg)
		if !isCall {
			return arg, chain, nil
		}
		fn, isTransform := transforms[name]
		if !isTransform {
			return "", nil, fmt.Errorf("%s(%s) is not a transform, so it cannot be an operand", name, strings.Join(args, ","))
		}
		if len(args) != 1 {
			return "", nil, fmt.Errorf("%s takes 1 arg, got %d", name, len(args))
		}
		chain = append(chain, fn)
		arg = args[0]
	}
}
func transformArg(fields map[string]node, a []string) error {
	leaf, _, err := unwrapTransform(a[0])
	if err != nil {
		return err
	}
	if isRef(leaf) {
		_, _, err := refShape(leaf)
		return err
	}
	return checkArm(leaf, fields, false)
}
func transformOperand(a []string) []string {
	leaf, _, err := unwrapTransform(a[0])
	if err != nil {
		return nil
	}
	return []string{leaf}
}
func transformPrep(outer func(string) string) func([]string) callFn {
	return func(a []string) callFn {
		_, chain, err := unwrapTransform(a[0])
		if err != nil {
			panic(fmt.Sprintf("fejkdata: transform arg %q reached prep unvalidated: %v", a[0], err))
		}
		return func(_ *session, _ string, operands []string) string {
			v := operands[0]
			for i := len(chain) - 1; i >= 0; i-- {
				v = chain[i](v)
			}
			return outer(v)
		}
	}
}

// asciiFolds maps the Latin letters with diacritics or ligatures to ASCII.
var asciiFolds = map[rune]string{
	'À': "A", 'Á': "A", 'Â': "A", 'Ã': "A", 'Ä': "A", 'Å': "A", 'Æ': "AE", 'Ç': "C",
	'È': "E", 'É': "E", 'Ê': "E", 'Ë': "E", 'Ì': "I", 'Í': "I", 'Î': "I", 'Ï': "I",
	'Ð': "D", 'Ñ': "N", 'Ò': "O", 'Ó': "O", 'Ô': "O", 'Õ': "O", 'Ö': "O", 'Ø': "O",
	'Ù': "U", 'Ú': "U", 'Û': "U", 'Ü': "U", 'Ý': "Y", 'Þ': "Th", 'ß': "ss", 'Œ': "OE",
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a", 'æ': "ae", 'ç': "c",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e", 'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'ð': "d", 'ñ': "n", 'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u", 'ý': "y", 'þ': "th", 'ÿ': "y", 'œ': "oe",
}

// asciiFold rewrites s to ASCII: folded Latin letters stay, any other non-ASCII
// rune is dropped.
func asciiFold(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r <= unicode.MaxASCII {
			b.WriteRune(r)
		} else {
			b.WriteString(asciiFolds[r])
		}
	}
	return b.String()
}
