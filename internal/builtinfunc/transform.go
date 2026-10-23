package builtinfunc

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// transforms are the builtins that rewrite one operand's value; they nest, so
// {lowercase(ascii(x))} folds then lowers.
var transforms = map[string]func(string) string{
	"ascii":     asciiFold,
	"lowercase": strings.ToLower,
	"slugify":   slugify,
	"uppercase": strings.ToUpper,
}

func withTransforms(calls map[string]builtin) map[string]builtin {
	for name, fn := range transforms {
		if _, clash := calls[name]; clash {
			panic(invariant.Broken("%s is registered as a builtin and a transform", name))
		}
		calls[name] = builtin{arity: 1, checkArgs: transformArg, prep: transformPrep(fn), operands: transformOperand}
	}
	return calls
}

// unwrapTransform peels nested transform calls off an operand arg, returning the
// field it finally names and the transforms to apply, innermost last.
func unwrapTransform(arg string) (leaf string, chain []func(string) string, err error) {
	for {
		name, args, isCall := grammar.FuncCall(arg)
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

// transformArg checks that each call nested in the arg is a transform taking one arg.
func transformArg(a []string) error {
	_, _, err := unwrapTransform(a[0])
	return err
}
func transformOperand(a []string) []string {
	if len(a) != 1 {
		return nil
	}
	leaf, _, err := unwrapTransform(a[0])
	if err != nil {
		return nil
	}
	return []string{leaf}
}
func transformPrep(outer func(string) string) func([]string) Call {
	return func(a []string) Call {
		_, chain, err := unwrapTransform(a[0])
		if err != nil {
			panic(invariant.Broken("transform arg %q reached prep unvalidated: %v", a[0], err))
		}
		return func(_ *drawstate.State, _ string, operands []string) string {
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
	'Ā': "A", 'ā': "a", 'Ă': "A", 'ă': "a", 'Ą': "A", 'ą': "a", 'Ć': "C",
	'ć': "c", 'Ĉ': "C", 'ĉ': "c", 'Ċ': "C", 'ċ': "c", 'Č': "C", 'č': "c", 'Ď': "D",
	'ď': "d", 'Đ': "D", 'đ': "d", 'Ē': "E", 'ē': "e", 'Ĕ': "E", 'ĕ': "e", 'Ė': "E",
	'ė': "e", 'Ę': "E", 'ę': "e", 'Ě': "E", 'ě': "e", 'Ĝ': "G", 'ĝ': "g", 'Ğ': "G",
	'ğ': "g", 'Ġ': "G", 'ġ': "g", 'Ģ': "G", 'ģ': "g", 'Ĥ': "H", 'ĥ': "h", 'Ħ': "H",
	'ħ': "h", 'Ĩ': "I", 'ĩ': "i", 'Ī': "I", 'ī': "i", 'Ĭ': "I", 'ĭ': "i", 'Į': "I",
	'į': "i", 'İ': "I", 'ı': "i", 'Ĳ': "IJ", 'ĳ': "ij", 'Ĵ': "J", 'ĵ': "j", 'Ķ': "K",
	'ķ': "k", 'ĸ': "k", 'Ĺ': "L", 'ĺ': "l", 'Ļ': "L", 'ļ': "l", 'Ľ': "L", 'ľ': "l",
	'Ŀ': "L", 'ŀ': "l", 'Ł': "L", 'ł': "l", 'Ń': "N", 'ń': "n", 'Ņ': "N", 'ņ': "n",
	'Ň': "N", 'ň': "n", 'Ŋ': "NG", 'ŋ': "ng", 'Ō': "O", 'ō': "o", 'Ŏ': "O", 'ŏ': "o",
	'Ő': "O", 'ő': "o", 'Ŕ': "R", 'ŕ': "r", 'Ŗ': "R", 'ŗ': "r", 'Ř': "R", 'ř': "r",
	'Ś': "S", 'ś': "s", 'Ŝ': "S", 'ŝ': "s", 'Ş': "S", 'ş': "s", 'Š': "S", 'š': "s",
	'Ţ': "T", 'ţ': "t", 'Ť': "T", 'ť': "t", 'Ŧ': "T", 'ŧ': "t", 'Ũ': "U", 'ũ': "u",
	'Ū': "U", 'ū': "u", 'Ŭ': "U", 'ŭ': "u", 'Ů': "U", 'ů': "u", 'Ű': "U", 'ű': "u",
	'Ų': "U", 'ų': "u", 'Ŵ': "W", 'ŵ': "w", 'Ŷ': "Y", 'ŷ': "y", 'Ÿ': "Y", 'Ź': "Z",
	'ź': "z", 'Ż': "Z", 'ż': "z", 'Ž': "Z", 'ž': "z", 'Ǫ': "O", 'ǫ': "o", 'Ǭ': "O",
	'ǭ': "o", 'ẞ': "SS",
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

func slugify(s string) string {
	var b strings.Builder
	dashed := true
	emit := func(c rune) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b.WriteRune(c)
			dashed = false
		case !dashed:
			b.WriteByte('-')
			dashed = true
		}
	}
	for _, r := range s {
		switch fold := asciiFolds[r]; {
		case r <= unicode.MaxASCII:
			emit(unicode.ToLower(r))
		case fold == "":
			emit('-')
		default:
			for _, c := range fold {
				emit(unicode.ToLower(c))
			}
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
