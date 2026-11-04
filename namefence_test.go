package fejkdata

import (
	"strings"
	"testing"
)

func loadFence(t *testing.T, files map[string]string) (*Generator, error) {
	t.Helper()
	return New(WithSeed(1), WithDataPath(writeData(t, files)))
}

func TestAReadThroughAFieldBoundNameIsRefusedWhereItsFieldReadsANameTheItemBinds(t *testing.T) {
	const ab = `{"format":"{written}","written":["x","y"],"legal":["L","M"]}`
	for label, c := range map[string]struct {
		co   string
		want []string
	}{
		"the field reads the name": {
			`{"format":"{registered as r}{r}","legal":"{r.legal}","registered":[{"format":"{/ab as w}{w}A","legal":"{w.legal}"},{"format":"{/ab as w}{w}B","legal":"{w.legal}"}]}`,
			[]string{"{r.legal}", `"w"`, `bind "w" outside the choice`}},
		"a field below it reads the name": {
			`{"format":"{registered as r}{r}","legal":"{r.legal}","registered":[{"format":"{/ab as w}{w}A","legal":{"format":"{x}","x":"{w.legal}"}},{"format":"{/ab as w}{w}B","legal":{"format":"{x}","x":"{w.legal}"}}]}`,
			[]string{"{r.legal}", `"w"`, `bind "w" outside the choice`}},
	} {
		t.Run(label, func(t *testing.T) {
			_, err := loadFence(t, map[string]string{"ab": ab, "co": c.co})
			if err == nil {
				t.Fatal("New() loaded, want it refused")
			}
			for _, want := range c.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("New() = %v, want an error naming %s", err, want)
				}
			}
		})
	}
}

func TestAReadThroughAFieldBoundNameLoadsWhereNoNameOfTheItemIsRead(t *testing.T) {
	const ab = `{"format":"{written}","written":["x","y"],"legal":["L","M"]}`
	for label, co := range map[string]string{
		"the name is bound outside the choice": `{"format":"{/ab as w}{registered as r}{r}","legal":"{r.legal}","registered":[{"format":"A","legal":"{w.legal}"},{"format":"B","legal":"{w.legal}"}]}`,
		"the field reads no name":              `{"format":"{registered as r}{r}","legal":"{r.legal}","registered":[{"format":"{/ab as w}{w}A","legal":"a"},{"format":"{/ab as w}{w}B","legal":"b"}]}`,
	} {
		t.Run(label, func(t *testing.T) {
			f, err := loadFence(t, map[string]string{"ab": ab, "co": co})
			if err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
			for i := 0; i < 20; i++ {
				fake(t, f, "co.legal")
			}
		})
	}
}

func TestAReadOfANameIsRefusedBesideAReferenceItsFormatMakesToThatPath(t *testing.T) {
	co := `{"format":"{/co.ab.name}","ab":{"format":"{name}","name":["a","b","c","d"]}}`
	_, err := loadFence(t, map[string]string{"co": co, "card": `"{/co as c}{c} {c.ab.name}"`})
	if err == nil {
		t.Fatal("New() loaded, want it refused")
	}
	for _, want := range []string{"{c}", "{/co.ab.name}", "{c.ab.name}"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("New() = %v, want an error naming %s", err, want)
		}
	}
}

func TestAReadOfANameLoadsBesideAReferenceThatReachesNoDrawOfItsPath(t *testing.T) {
	for label, c := range map[string]struct {
		files map[string]string
		want  string
	}{
		"a reference to another category": {map[string]string{
			"other": `{"format":"{name}","name":["a","b"]}`,
			"co":    `{"format":"{/other.name}","ab":{"format":"{name}","name":["c","d"]}}`,
			"card":  `"{/co as c}{c} {c.ab.name}"`,
		}, ""},
		"a reference to a constant": {map[string]string{
			"co":   `{"format":"{/co.ab.name}","ab":{"format":"{name}","name":"k"}}`,
			"card": `"{/co as c}{c} {c.ab.name}"`,
		}, "k k"},
		"a path above the reference's": {map[string]string{
			"co":   `{"format":"{/co.ab.name}","ab":{"format":"{name}","name":["a","b"]}}`,
			"card": `"{/co as c}{c} {c.ab}"`,
		}, ""},
	} {
		t.Run(label, func(t *testing.T) {
			f, err := loadFence(t, c.files)
			if err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
			if got := fake(t, f, "card"); c.want != "" && got != c.want {
				t.Errorf("card = %q, want %q", got, c.want)
			}
		})
	}
}

func TestReadsOfANameLoadWhereTheirSharedLevelsDrawNothingThatCouldDiffer(t *testing.T) {
	for label, c := range map[string]struct{ cat, card, wantFirst string }{
		"two fields of a plain template, and a third beside them": {
			`{"format":"{w.a} {w.b}","w":{"format":"{a}","a":["x"],"b":["p"],"z":["m","n"]}}`,
			`"{/cat as n}{n}|{n.w.z}"`, "x p"},
		"a constant read twice": {
			`{"format":"{w}-{w}","w":"k"}`,
			`"{/cat as n}{n}|{n.w}"`, "k-k"},
		"a constant bound to a name": {
			`{"format":"{w as r}{r}","w":"k"}`,
			`"{/cat as n}{n}|{n.w}"`, "k"},
		"a template and a field of it": {
			`{"format":"{w} {w.z}","w":{"format":"{a}","a":["x"],"z":["m"]}}`,
			`"{/cat as n}{n}|{n.w.z}"`, "x m"},
		"a plain template's field bound to a name": {
			`{"format":"{a.b as r}{r}","a":{"format":"{b}","b":["x"],"z":["m","n"]}}`,
			`"{/cat as n}{n}|{n.a.z}"`, "x"},
	} {
		t.Run(label, func(t *testing.T) {
			f, err := loadFence(t, map[string]string{"cat": c.cat, "card": c.card})
			if err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
			if got, _, _ := strings.Cut(fake(t, f, "card"), "|"); got != c.wantFirst {
				t.Errorf("card = %q, want it to start %q", got, c.wantFirst)
			}
		})
	}
}

func TestReadsOfANameAreRefusedWhereASharedChoiceCouldDrawApart(t *testing.T) {
	item := `{"format":"{a}","a":["x","y"],"b":["p","q"],"z":["m","n"]}`
	for label, c := range map[string]struct{ cat, card, want string }{
		"two fields of a choice": {
			`{"format":"{w.a} {w.b}","w":[` + item + `,` + item + `]}`,
			`"{/cat as n}{n}|{n.w.z}"`, `renders field "w" twice`},
		"a field of a choice bound to a name": {
			`{"format":"{a.b as r}{r}","a":[` + item + `,` + item + `]}`,
			`"{/cat as n}{n}|{n.a.z}"`, `through name "r"`},
		"a plain template read twice": {
			`{"format":"{w}-{w}","w":{"format":"{a}","a":["x","y"]}}`,
			`"{/cat as n}{n}|{n.w}"`, `renders field "w" twice`},
	} {
		t.Run(label, func(t *testing.T) {
			_, err := loadFence(t, map[string]string{"cat": c.cat, "card": c.card})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("New() = %v, want an error holding %q", err, c.want)
			}
		})
	}
}
