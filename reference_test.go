package fejkdata

import (
	"strings"
	"testing"
)

// TestRootReferenceAcrossFolders is the headline case: a category in one folder
// pulls a value from another via a {/path} reference resolved from the data root.
func TestRootReferenceAcrossFolders(t *testing.T) {
	dir := writeData(t, map[string]string{
		"en_US/person":   `"Pat Smith"`,
		"sv_SE/greeting": `"Hej, {/en_US.person}!"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if got := fake(t, f, "sv_SE.greeting"); got != "Hej, Pat Smith!" {
		t.Fatalf("greeting = %q, want \"Hej, Pat Smith!\"", got)
	}
}

// TestReferenceIntoAField reaches a field inside a referenced category.
func TestReferenceIntoAField(t *testing.T) {
	dir := writeData(t, map[string]string{
		"who":  `{"format":"{first} {last}","first":"Ada","last":"Byron"}`,
		"card": `"signed {/who.last}"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if got := fake(t, f, "card"); got != "signed Byron" {
		t.Fatalf("card = %q, want \"signed Byron\"", got)
	}
}

// TestReferenceInAlternation lets a reference stand as one arm of a {a|/b}
// alternation, so a field and a cross-file value share one slot.
func TestReferenceInAlternation(t *testing.T) {
	dir := writeData(t, map[string]string{
		"far":  `"X"`,
		"near": `{"format":"{here|/far}","here":"H"}`,
	})
	f := newGenerator(t, dir, WithSeed(2))
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		seen[fake(t, f, "near")] = true
	}
	if !seen["H"] || !seen["X"] || len(seen) != 2 {
		t.Fatalf("alternation produced %v, want both H and X", seen)
	}
}

// TestReferenceCombinesLoadedPaths is the point of references over the merge
// model: data layered from two dirs can point at each other through the root.
func TestReferenceCombinesLoadedPaths(t *testing.T) {
	a := writeData(t, map[string]string{"en_US/word": `"river"`})
	b := writeData(t, map[string]string{"mine/slug": `"the-{/en_US.word}"`})
	f := newGeneratorN(t, []string{a, b}, WithSeed(1))
	if got := fake(t, f, "mine.slug"); got != "the-river" {
		t.Fatalf("slug = %q, want the-river", got)
	}
}

// TestReferenceChain follows a reference to a node that is itself a reference, so
// linking order cannot matter.
func TestReferenceChain(t *testing.T) {
	dir := writeData(t, map[string]string{
		"a": `"{/b}"`,
		"b": `"{/c}"`,
		"c": `"deep"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if got := fake(t, f, "a"); got != "deep" {
		t.Fatalf("a = %q, want deep", got)
	}
}

// TestReferenceErrors lists the references New must reject up front, so a bad path
// fails at load, never at a random render.
func TestReferenceErrors(t *testing.T) {
	cases := map[string]map[string]string{
		"missing target": {"card": `"{/nope.gone}"`},
		"folder target":  {"en_US/word": `"w"`, "card": `"{/en_US}"`},
		"a variant on the path lacks the field": {
			"who":  `[{"format":"{f}","f":"1"},{"format":"{g}","g":"2"}]`,
			"card": `"{/who.f}"`,
		},
		"empty reference path": {"card": `"{/}"`},
		// A reference that leads back to its own value never terminates at render,
		// so New must reject the cycle up front (direct, mutual, or chained).
		"direct cycle": {"a": `"x{/a}"`},
		"mutual cycle": {"a": `"{/b}"`, "b": `"{/a}"`},
		"chain cycle":  {"a": `"{/b}"`, "b": `"{/c}"`, "c": `"{/a}"`},
		// calc renders its operands, so a cycle through one must be caught too.
		"calc operand cycle": {"x": `{"format":"{calc(y)}","y":"{/x}"}`},
		// A field its parent's format never renders is still reachable by dot path,
		// so a cycle hiding in one must fail at New rather than at render.
		"cycle in an unrendered field": {"cat": `{"format":"hi","x":"{/cat.x}"}`},
		"mutual cycle between unrendered fields": {
			"cat": `{"format":"hi","x":"{/cat.y}","y":"{/cat.x}"}`,
		},
		"cycle in an unrendered field of a choice arm": {
			"cat": `{"format":"hi","x":"{/cat.x}"}`,
		},
		// The shipped layout puts categories in folders, so a cycle one level down
		// is the common case, not an edge case.
		"cycle in a subfolder":            {"sv_SE/a": `"x{/sv_SE.a}"`},
		"mutual cycle within a subfolder": {"sv_SE/a": `"{/sv_SE.b}"`, "sv_SE/b": `"{/sv_SE.a}"`},
		"mutual cycle across two folders": {"en_US/a": `"{/sv_SE.b}"`, "sv_SE/b": `"{/en_US.a}"`},
		// ".." is reserved for bound references, so an authored key using it would
		// name a node nothing can reach and nothing would validate.
		"field key using the reference prefix": {"cat": `{"format":"hi","..x":"{/nope}"}`},
	}
	for name, files := range cases {
		if _, err := New(WithoutShippedData(), WithDataPath(writeData(t, files))); err == nil {
			t.Errorf("%s: New = nil error, want a reference error", name)
		}
	}
}

// TestDotPrefixedDataEntriesAreSkipped pins what a leading dot means on disk: the
// entry is hidden, not data, so a data directory can also be a checkout. A name
// starting with the reference prefix is covered by that same rule, since ".."
// starts with "." — it is skipped, not rejected.
func TestDotPrefixedDataEntriesAreSkipped(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{
		"sv_SE/ok":     `"fine"`,
		"sv_SE/..bad":  `"{/nope}"`,
		"sv_SE/..y/ct": `"{/nope}"`,
		".git/config":  `"not data"`,
	}), WithSeed(1))
	if got := f.List(); len(got) != 1 || got[0] != "sv_SE.ok" {
		t.Fatalf("List() = %v, want only sv_SE.ok", got)
	}
}

// TestReferenceFromUnrenderedFieldTerminates guards the cycle check against
// over-rejecting: a field the format never renders may point back at its own
// category, which terminates, and stays renderable by path.
func TestReferenceFromUnrenderedFieldTerminates(t *testing.T) {
	dir := writeData(t, map[string]string{"cat": `{"format":"hi","x":"see {/cat}"}`})
	f := newGenerator(t, dir, WithSeed(1))
	if got := fake(t, f, "cat"); got != "hi" {
		t.Fatalf("cat = %q, want hi", got)
	}
	if got := fake(t, f, "cat.x"); got != "see hi" {
		t.Fatalf("cat.x = %q, want \"see hi\"", got)
	}
}

// TestNewErrorIsDeterministic pins one message per broken data set: map iteration
// order must not decide which of several problems the user is told about, whether
// they sit in separate categories or in one template's fields.
func TestNewErrorIsDeterministic(t *testing.T) {
	cases := map[string]map[string]string{
		"three bad references": {
			"a": `"{/nope.one}"`,
			"b": `"{/nope.two}"`,
			"c": `"{/nope.three}"`,
		},
		"two bad fields in one template": {
			"cat": `{"format":"hi","aaa":{"no":1},"zzz":{"no":2}}`,
		},
	}
	for name, files := range cases {
		dir := writeData(t, files)
		var first string
		for i := 0; i < 50; i++ {
			_, err := New(WithoutShippedData(), WithDataPath(dir))
			if err == nil {
				t.Fatalf("%s: New = nil error, want a load error", name)
			}
			if i == 0 {
				first = err.Error()
				continue
			}
			if err.Error() != first {
				t.Fatalf("%s: New error varies between runs:\n  %s\n  %s", name, first, err.Error())
			}
		}
		t.Logf("%s -> %s", name, first)
	}
}

// TestNewErrorPathIsCanonical pins the node path a load error names: a choice arm
// adds no segment, and a bound {/path} reference is not a containment segment at
// all, so a bad reference is reported against the node that holds it.
func TestNewErrorPathIsCanonical(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			"cycle inside a choice arm",
			map[string]string{"cat": `{"format":"hi","x":"{/cat.x}"}`},
			"fejkdata: reference cycle: cat.x -> /cat.x",
		},
		{
			"bad reference reached through another reference",
			map[string]string{"a": `"{/b}"`, "b": `"{/nope}"`},
			`fejkdata: b: reference {/nope}: no entry "nope"`,
		},
	}
	for _, c := range cases {
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, c.files)))
		if err == nil {
			t.Errorf("%s: New = nil error", c.name)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("%s:\n  got  %s\n  want %s", c.name, err, c.want)
		}
	}
}

func TestReferencePathIsHeld(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `[{"format":"{first} {last}","first":"Anna","last":"Andersson"},{"format":"{first} {last}","first":"Bo","last":"Berg"}]`,
		"card":   `"{/person.first} {/person.last}"`,
	})
	f := newGenerator(t, dir, WithSeed(3))
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		got := fake(t, f, "card")
		if got != "Anna Andersson" && got != "Bo Berg" {
			t.Fatalf("card = %q, want one person's first and last name", got)
		}
		seen[got] = true
	}
	if len(seen) != 2 {
		t.Fatalf("card only ever rendered %v", seen)
	}
}

func TestReferenceThroughChoiceNeedsEveryVariant(t *testing.T) {
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"who":  `[{"format":"{f}{h}","f":"1","h":"x"},{"format":"{g}{h}","g":"2","h":"y"}]`,
		"card": `"{/who.f}"`,
	})))
	if err == nil || !strings.Contains(err.Error(), "not every variant") {
		t.Fatalf("New = %v, want the missing variant named", err)
	}
}

func TestBareReferenceDrawsEachTime(t *testing.T) {
	dir := writeData(t, map[string]string{
		"die":  `["1","2","3","4","5","6"]`,
		"roll": `"{/die} {/die}"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "roll"); got[0] != got[2] {
			return
		}
	}
	t.Fatal("two bare references always agreed; each should be its own draw")
}

func TestReferenceOverlapIsRejected(t *testing.T) {
	p := `[{"format":"{first}","first":"A","last":"1"},{"format":"{first}","first":"B","last":"2"}]`
	for name, file := range map[string]string{
		"head beside a path":                                        `{"format":"{/cat.p} {/cat.p.first}","p":` + p + `}`,
		"sibling path beside a reference path":                      `{"format":"{p.first} {/cat.p.last}","p":` + p + `}`,
		"sibling fields reading a level and a path into it":         `{"format":"{a} {b}","a":"{/cat.p}","b":"{/cat.p.first}","p":` + p + `}`,
		"a field rendering the level a nested reference reads into": `{"format":"{x} {p}","x":"{/cat.p.first}","p":` + p + `}`,
		"a bare reference beside a path into what it never renders": `{"format":"{a} {b}","a":"{/other}","b":"{/other.p.first}"}`,
	} {
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"cat": file, "other": `{"format":"x","p":` + p + `}`})))
		if err == nil || !strings.Contains(err.Error(), "reads a path into") {
			t.Errorf("%s: New = %v, want the overlap rejected", name, err)
		}
	}
	if _, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"cat": `{"format":"{a} {b}","a":{"format":"{/cat.p}","drawGroup":"g"},"b":"{/cat.p.first}","p":` + p + `}`,
	}))); err != nil {
		t.Errorf("New = %v, want a level and a path into it accepted in groups of their own", err)
	}
}

const drawPeople = `[{"format":"{first} {last}","first":"Ada","last":"Lovelace"},{"format":"{first} {last}","first":"Bo","last":"Ek"},{"format":"{first} {last}","first":"Cy","last":"Young"}]`

// onePerson reports whether name is the first name and surname of one drawPeople row.
func onePerson(name string) bool {
	first, last, _ := strings.Cut(name, " ")
	return last != "" && map[string]string{"Ada": "Lovelace", "Bo": "Ek", "Cy": "Young"}[first] == last
}

func TestAReferencePathIsOneDrawPerRender(t *testing.T) {
	dir := writeData(t, map[string]string{
		"apart":      `{"format":"{a} & {b}","a":{"format":"{/person.first} {/person.last}","drawGroup":"x"},"b":{"format":"{/person.first} {/person.last}","drawGroup":"y"}}`,
		"caller":     `{"format":"{a} & {b}","a":{"format":"{/person.first} {/person.last}","drawGroup":"x"},"b":"{/pay}"}`,
		"contact":    `{"format":"{first} {last} <{email}>","email":"{lowercase(/person.first)}.{lowercase(/person.last)}@example.com","first":"{/person.first}","last":"{/person.last}"}`,
		"iterations": `{"format":"{/person.first} {r}","drawGroup":"outer","r":{"format":"{a}={b}","repeat":3,"separator":",","a":"{/person.first}","b":{"format":"{/person.first}","drawGroup":"outer"}}}`,
		"nested":     `{"format":"{/person.first} {inner}","inner":"{/person.last}"}`,
		"pair":       `{"format":"{a} & {b}","a":"{/person.first} {/person.last}","b":"{/person.first} {/person.last}"}`,
		"pay":        `{"format":"{p}","p":{"format":"{/person.first} {/person.last}","drawGroup":"x"}}`,
		"person":     drawPeople,
	})
	f := newGenerator(t, dir, WithSeed(1))
	apart, local, noGroup := false, false, false
	for i := 0; i < 100; i++ {
		a, b, _ := strings.Cut(fake(t, f, "caller"), " & ")
		if !onePerson(a) || !onePerson(b) {
			t.Fatalf("caller = %q & %q, want each one person", a, b)
		}
		local = local || a != b
		_, iterations, _ := strings.Cut(fake(t, f, "iterations"), " ")
		for _, pair := range strings.Split(iterations, ",") {
			a, b, _ := strings.Cut(pair, "=")
			noGroup = noGroup || a != b
		}
		if got := fake(t, f, "nested"); !onePerson(got) {
			t.Fatalf("nested = %q, want a nested template's path one draw with its parent's", got)
		}
		name, email, _ := strings.Cut(fake(t, f, "contact"), " <")
		first, last, _ := strings.Cut(name, " ")
		if !onePerson(name) || email != strings.ToLower(first)+"."+strings.ToLower(last)+"@example.com>" {
			t.Fatalf("contact = %q <%s, want first, last and email one person", name, email)
		}
		r, err := f.FakeRecord("contact")
		if err != nil {
			t.Fatal(err)
		}
		c := r.Columns()
		if !onePerson(c[1].Value+" "+c[2].Value) || c[0].Value != strings.ToLower(c[1].Value)+"."+strings.ToLower(c[2].Value)+"@example.com" {
			t.Fatalf("contact record %s, want first, last and email one person", r.JSON())
		}
		a, b, _ = strings.Cut(fake(t, f, "pair"), " & ")
		if !onePerson(a) || a != b {
			t.Fatalf("pair = %q & %q, want both fields one person", a, b)
		}
		a, b, _ = strings.Cut(fake(t, f, "apart"), " & ")
		if !onePerson(a) || !onePerson(b) {
			t.Fatalf("apart = %q & %q, want each group one person", a, b)
		}
		apart = apart || a != b
	}
	if !apart {
		t.Error("groups x and y drew one person in 100 renders, want a draw each")
	}
	if !local {
		t.Error("group x in caller and group x in the pay it references drew one person in 100 renders; a group name is local to its category")
	}
	if !noGroup {
		t.Error("an iteration's plain read and its group outer always agreed; a repeat iteration renders in no group, so they are a draw each")
	}
}

func TestRelativeReferences(t *testing.T) {
	dir := writeData(t, map[string]string{
		"sv_SE/username":  `"bob"`,
		"sv_SE/email":     `"{.username}@example.com"`,
		"sv_SE/deep/card": `"{..username} via {/sv_SE.username}"`,
		"top":             `"{.sv_SE.username}"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for path, want := range map[string]string{
		"sv_SE.email":     "bob@example.com",
		"sv_SE.deep.card": "bob via bob",
		"top":             "bob",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
}

func TestRelativeAndRootSpellingsBindOneDraw(t *testing.T) {
	dir := writeData(t, map[string]string{
		"sv_SE/person": `[{"format":"{first} {last}","first":"Anna","last":"Andersson"},{"format":"{first} {last}","first":"Bo","last":"Berg"}]`,
		"sv_SE/card":   `"{.person.first} {/sv_SE.person.last}"`,
	})
	f := newGenerator(t, dir, WithSeed(3))
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "sv_SE.card"); got != "Anna Andersson" && got != "Bo Berg" {
			t.Fatalf("card = %q, want both spellings to read one person", got)
		}
	}
}

func TestReferenceSigilErrors(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"no folder above the root": {"a": `"{..b}"`, "b": `"x"`},
		"three dots":               {"a": `"{...b}"`, "b": `"x"`},
		"root sigil alone":         {"a": `"{/}"`},
		"dot alone":                {"a": `"{.}"`},
		"missing sibling":          {"sv_SE/a": `"{.nope}"`},
		"slash in a field name":    {"a": `{"format":"{x}","x":"1","a/b":"2"}`},
	} {
		if _, err := New(WithoutShippedData(), WithDataPath(writeData(t, files))); err == nil {
			t.Errorf("%s: New = nil error, want a reference error", name)
		}
	}
}

func TestSpellingsOfOneReferenceAreOneLevel(t *testing.T) {
	person := `[{"format":"{first} {last}","first":"Ada","last":"Byron"},{"format":"{first} {last}","first":"Bo","last":"Ek"}]`
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"sv_SE/person": person,
		"sv_SE/mail":   `"{.person} <{/sv_SE.person.first}>"`,
	})))
	if err == nil || !strings.Contains(err.Error(), "reads a path into") || !strings.Contains(err.Error(), "{.person}") {
		t.Errorf("New = %v, want the bare spelling rejected beside the path spelling", err)
	}
	dir := writeData(t, map[string]string{
		"sv_SE/word": `["alpha","beta","gamma"]`,
		"sv_SE/loud": `"{/sv_SE.word} {uppercase(.word)}"`,
	})
	f := newGenerator(t, dir, WithSeed(2))
	for i := 0; i < 50; i++ {
		got := strings.Fields(fake(t, f, "sv_SE.loud"))
		if len(got) != 2 || strings.ToUpper(got[0]) != got[1] {
			t.Fatalf("loud = %q, want one draw under both spellings", got)
		}
	}
}

func TestSlashAfterAReferenceSigilIsRejected(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"after ..": {"sv_SE/person": `"Ada"`, "sv_SE/deep/a": `"{../person}"`},
		"after .":  {"sv_SE/person": `"Ada"`, "sv_SE/a": `"{./person}"`},
	} {
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, files)))
		if err == nil || !strings.Contains(err.Error(), "person}") || !strings.Contains(err.Error(), "write {") {
			t.Errorf("%s: New = %v, want the slash rejected naming the spelling", name, err)
		}
	}
}
