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
// resolve order cannot matter.
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
		// A reference that leads back to its own value never terminates at render, so
		// New must reject the cycle up front (mutual or chained). One into its own
		// category is refused before the cycle walk reaches it, as a unit rule.
		"a category referencing itself": {"a": `"x{/a}"`},
		"mutual cycle":                  {"a": `"{/b}"`, "b": `"{/a}"`},
		"chain cycle":                   {"a": `"{/b}"`, "b": `"{/c}"`, "c": `"{/a}"`},
		// calc renders its operands, so a reference through one is caught too.
		"a calc operand into its own category": {"x": `{"format":"{calc(y)}","y":"{/x}"}`},
		// A field its parent's format never renders is still reachable by dot path,
		// so what hides in one must fail at New rather than at render.
		"an unrendered field into its own category": {"cat": `{"format":"hi","x":"{/cat.x}"}`},
		"two unrendered fields into their own category": {
			"cat": `{"format":"hi","x":"{/cat.y}","y":"{/cat.x}"}`,
		},
		// The shipped layout puts categories in folders, so one level down is the
		// common case, not an edge case.
		"a category in a subfolder referencing itself": {"sv_SE/a": `"x{/sv_SE.a}"`},
		"mutual cycle within a subfolder":              {"sv_SE/a": `"{/sv_SE.b}"`, "sv_SE/b": `"{/sv_SE.a}"`},
		"mutual cycle across two folders":              {"en_US/a": `"{/sv_SE.b}"`, "sv_SE/b": `"{/en_US.a}"`},
		// ".." is reserved for references, so an authored key using it would
		// name a node nothing can reach and nothing would validate.
		"field key using the reference prefix": {"cat": `{"format":"hi","..x":"{/nope}"}`},
	}
	for name, files := range cases {
		if _, err := New(WithDataPath(writeData(t, files))); err == nil {
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

func TestAReferenceIntoItsOwnCategoryIsAFreshDraw(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"cat": `{"format":"{y} {x}","x":"{/cat.y}","y":["1","2"]}`}), WithSeed(1))
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		seen[fake(t, f, "cat")] = true
	}
	if !seen["1 2"] && !seen["2 1"] {
		t.Errorf("cat drew %v in 50 renders, want x a draw apart from y", seen)
	}
	for name, file := range map[string]string{
		"the category in its format": `{"format":"{/cat}"}`,
		"a field reading itself":     `{"format":"{x}","x":"{/cat.x}"}`,
	} {
		_, err := New(WithDataPath(writeData(t, map[string]string{"cat": file})))
		if err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Errorf("%s: New = %v, want the cycle refused", name, err)
		}
	}
	// Reading a sibling as a path is the spelling that stays.
	if _, err := New(WithDataPath(writeData(t, map[string]string{
		"cat": `{"format":"{y.v}","y":{"format":"{v}","v":["1","2"]}}`,
	}))); err != nil {
		t.Errorf("New = %v, want the sibling path accepted", err)
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
			_, err := New(WithDataPath(dir))
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
// adds no segment, and a {/path} reference is not a containment segment at
// all, so a bad reference is reported against the node that holds it.
func TestNewErrorPathIsCanonical(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			"cycle through another category",
			map[string]string{"cat": `{"format":"hi","x":"{/hop}"}`, "hop": `"{/cat.x}"`},
			"fejkdata: reference cycle: cat.x -> /hop -> /cat.x",
		},
		{
			"bad reference reached through another reference",
			map[string]string{"a": `"{/b}"`, "b": `"{/nope}"`},
			`fejkdata: b: reference {/nope}: no entry "nope", and the module holding b names no module it reads by default`,
		},
	}
	for _, c := range cases {
		_, err := New(WithDataPath(writeData(t, c.files)))
		if err == nil {
			t.Errorf("%s: New = nil error", c.name)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("%s:\n  got  %s\n  want %s", c.name, err, c.want)
		}
	}
}

func TestReferenceThroughChoiceNeedsEveryVariant(t *testing.T) {
	_, err := New(WithDataPath(writeData(t, map[string]string{
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

func TestReferencesIntoOneLevelLoad(t *testing.T) {
	cat := `{"format":"x","p":[{"format":"{first}","first":"A","last":"1"},{"format":"{first}","first":"B","last":"2"}]}`
	for _, files := range []map[string]string{
		{"cat": cat, "row": `"{/cat.p} {/cat.p.first}"`},
		{"cat": cat, "row": `{"format":"{a} {b}","a":"{/cat.p}","b":"{/cat.p.first}"}`},
		{"cat": cat, "row": `{"format":"{x} {p}","x":"{/cat.p.first}","p":"{/cat.p}"}`},
		{"cat": cat, "row": `{"format":"{a} {b}","a":"{/cat}","b":"{/cat.p.first}"}`},
		{"row": `{"format":"{p.first} {/hop}","p":[{"format":"{first}","first":"A","last":"1"},{"format":"{first}","first":"B","last":"2"}]}`, "hop": `"{/row.p.last}"`},
		{"sv_SE/person": cat, "sv_SE/mail": `"{.person} <{/sv_SE.person.p.first}>"`},
	} {
		f, err := New(WithDataPath(writeData(t, files)), WithSeed(1))
		if err != nil {
			t.Errorf("New(%v) = %v, want each read a draw of its own", files, err)
			continue
		}
		for path := range files {
			fake(t, f, strings.ReplaceAll(path, "/", "."))
		}
	}
}

const drawPeople = `[{"format":"{first} {last}","first":"Ada","last":"Lovelace"},{"format":"{first} {last}","first":"Bo","last":"Ek"},{"format":"{first} {last}","first":"Cy","last":"Young","born":"1867"}]`

// onePerson reports whether name is the first name and surname of one drawPeople row.
func onePerson(name string) bool {
	first, last, _ := strings.Cut(name, " ")
	return last != "" && map[string]string{"Ada": "Lovelace", "Bo": "Ek", "Cy": "Young"}[first] == last
}

func TestFieldsReadOneNameTheirCategoryBinds(t *testing.T) {
	dir := writeData(t, map[string]string{
		"contact": `{"format":"{first} {last} <{email}>","email":"{lowercase(p.first)}.{lowercase(p.last)}@example.com","first":"{/person as p}{p.first}","last":"{p.last}"}`,
		"person":  drawPeople,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for i := 0; i < 100; i++ {
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

func TestReferenceSigilErrors(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"no folder above the root": {"a": `"{..b}"`, "b": `"x"`},
		"three dots":               {"a": `"{...b}"`, "b": `"x"`},
		"root sigil alone":         {"a": `"{/}"`},
		"dot alone":                {"a": `"{.}"`},
		"missing sibling":          {"sv_SE/a": `"{.nope}"`},
		"slash in a field name":    {"a": `{"format":"{x}","x":"1","a/b":"2"}`},
	} {
		if _, err := New(WithDataPath(writeData(t, files))); err == nil {
			t.Errorf("%s: New = nil error, want a reference error", name)
		}
	}
}

func TestSlashAfterAReferenceSigilIsRejected(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"after ..": {"sv_SE/person": `"Ada"`, "sv_SE/deep/a": `"{../person}"`},
		"after .":  {"sv_SE/person": `"Ada"`, "sv_SE/a": `"{./person}"`},
	} {
		_, err := New(WithDataPath(writeData(t, files)))
		if err == nil || !strings.Contains(err.Error(), "person}") || !strings.Contains(err.Error(), "write {") {
			t.Errorf("%s: New = %v, want the slash rejected naming the spelling", name, err)
		}
	}
}

func TestACycleThroughASelectedRowIsRefused(t *testing.T) {
	_, err := New(WithDataPath(writeFiles(t, map[string]string{
		"x.json": `{"format":"{v} {/y}","rows":"x.tsv","key":"code"}`,
		"x.tsv":  "code\tv\n1\ta\n2\tb\n",
		"y.json": `"{/x[1]}"`,
	})))
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("New = %v, want the cycle through x[1] refused", err)
	}
}
