package fejkdata

import (
	"strings"
	"testing"
)

func loadFence(t *testing.T, files map[string]string) (*Generator, error) {
	t.Helper()
	return New(WithSeed(1), onTestDay, WithDataPath(writeData(t, files)))
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

func TestANameReadInsideTheItemThatBindsItLoads(t *testing.T) {
	const ab = `{"format":"{written}","written":["x","y"],"legal":["L","M"]}`
	for label, co := range map[string]string{
		"a name bound to a field of the item": `{"format":"{registered}","registered":[{"format":"{/ab as w}{w}{legal as r} {r}","legal":"{w.legal}"},{"format":"{/ab as w}{w}{legal as r} {r}!","legal":"{w.legal}"}]}`,
		"a read of a field of that field":     `{"format":"{registered}","registered":[{"format":"{/ab as w}{w}{inner as r} {r.legal}","inner":{"format":"a","legal":"{w.legal}"}},{"format":"{/ab as w}{w}{inner as r} {r.legal}!","inner":{"format":"b","legal":"{w.legal}"}}]}`,
	} {
		t.Run(label, func(t *testing.T) {
			f, err := loadFence(t, map[string]string{"ab": ab, "co": co})
			if err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
			for i := 0; i < 20; i++ {
				fake(t, f, "co")
			}
		})
	}
}

func TestAChoiceItemThatReadsOnlyANameItBindsDraws(t *testing.T) {
	const ab = `{"format":"{written}","written":["x","y"],"legal":["L","M"]}`
	const item = `[{"format":"{/ab as w}{w.legal}"},{"format":"{/ab as w}{w.legal}!"}]`
	for label, c := range map[string]struct{ co, path string }{
		"the choice is the category's format":  {`{"format":"{legal}","legal":` + item + `}`, "co"},
		"the choice is read through a field":   {`{"format":"{legal}","legal":` + item + `}`, "co.legal"},
		"the choice is a field of a bound row": {`{"format":"{p as r}{r.legal}","p":{"format":"a","legal":` + item + `}}`, "co"},
	} {
		t.Run(label, func(t *testing.T) {
			f, err := loadFence(t, map[string]string{"ab": ab, "co": c.co})
			if err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
			for i := 0; i < 40; i++ {
				switch got := fake(t, f, c.path); got {
				case "L", "M", "L!", "M!":
				default:
					t.Fatalf("draw %d = %q, want L, M, L! or M!", i, got)
				}
			}
		})
	}
}

func TestAReadOfANameIsRefusedBesideAReferenceItsFormatMakesToThatPath(t *testing.T) {
	for label, c := range map[string]struct {
		files map[string]string
		want  []string
	}{
		"a reference to a field": {map[string]string{
			"co":   `{"format":"{/co.ab.name}","ab":{"format":"{name}","name":["a","b","c","d"]}}`,
			"card": `"{/co as c}{c} {c.ab.name}"`,
		}, []string{"{c}", "{/co.ab.name}", "{c.ab.name}"}},
		"a path through a choice to constants": {map[string]string{
			"co":   `{"format":"{/co.w.name}","w":[{"format":"A","name":"a"},{"format":"B","name":"b"}]}`,
			"card": `"{/co as c}{c} {c.w.name}"`,
		}, []string{"{c}", "{/co.w.name}", "{c.w.name}"}},
		"a reference reached through another category": {map[string]string{
			"other": `"{/co.ab.name}"`,
			"co":    `{"format":"{/other}","ab":{"format":"{name}","name":["a","b","c","d"]}}`,
			"card":  `"{/co as c}{c} {c.ab.name}"`,
		}, []string{"{c}", "{/co.ab.name}", "{c.ab.name}"}},
		"a name bound to a field of the category": {map[string]string{
			"co":   `{"format":"{w}","w":{"format":"{/co.w.a}","a":["x","y","z","q"]}}`,
			"card": `"{/co.w as n}{n}|{n.a}"`,
		}, []string{"{n}", "{/co.w.a}", "{n.a}"}},
		"a path three fields down beside siblings": {map[string]string{
			"co":   `{"format":"{/co.a.b.c}","a":{"format":"{b.c}","e":"s","b":{"format":"{c}","d":"t","c":["p","q"]}}}`,
			"card": `"{/co as c}{c} {c.a.b.c}"`,
		}, []string{"{c}", "{/co.a.b.c}", "{c.a.b.c}"}},
	} {
		t.Run(label, func(t *testing.T) {
			_, err := loadFence(t, c.files)
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

func TestAReadOfANameIsRefusedBesideAReadThatDrawsAfreshAFieldItReadsFromThePick(t *testing.T) {
	for label, c := range map[string]struct {
		files map[string]string
		want  []string
	}{
		"a field of a field": {map[string]string{
			"cat":  `{"format":"{w.a}","w":{"format":"{a}","a":["x","y","z","q"],"b":["p"]}}`,
			"card": `"{/cat as n}{n}|{n.w}"`,
		}, []string{"{n}", "{w.a}", "{n.w}", "{a}", "{n.w.a}", "{/cat.w}"}},
		"the read of the field comes first": {map[string]string{
			"cat":  `{"format":"{a.b}","a":{"format":"{b}","b":["x","y"]}}`,
			"card": `"{/cat as n}{n.a}|{n}"`,
		}, []string{"{n}", "{a.b}", "{n.a}", "{b}", "{n.a.b}"}},
		"both end on a template that isn't a constant": {map[string]string{
			"cat":  `{"format":"{w.a}","w":{"format":"{a}","a":{"format":"{b}","b":["x","y","z","q"]}}}`,
			"card": `"{/cat as n}{n}|{n.w}"`,
		}, []string{"{n}", "{w.a}", "{n.w}", "{a}", "{n.w.a}"}},
		"the read from the pick goes deeper than the fresh one": {map[string]string{
			"cat":  `{"format":"{w.a.b}","w":{"format":"{a}","a":{"format":"{b}","b":["x","y","z","q"]}}}`,
			"card": `"{/cat as n}{n}|{n.w}"`,
		}, []string{"{n}", "{w.a.b}", "{n.w}", "{a}", "{n.w.a}"}},
		"the fresh read goes deeper than the one from the pick": {map[string]string{
			"cat":  `{"format":"{w.a}","w":{"format":"{a.b}","a":{"format":"{b}","b":["x","y","z","q"]}}}`,
			"card": `"{/cat as n}{n}|{n.w}"`,
		}, []string{"{n}", "{w.a}", "{n.w}", "{a.b}", "{n.w.a}"}},
	} {
		t.Run(label, func(t *testing.T) {
			_, err := loadFence(t, c.files)
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

func TestAReadOfANameLoadsBesideAReadWhoseFormatReadsOneFieldOneWay(t *testing.T) {
	const cat = `{"format":"{w.a}","w":{"format":"{a}","a":["x","y","z","q"],"b":["p"]}}`
	for label, c := range map[string]struct{ cat, card string }{
		"the field is constant":                  {`{"format":"{w.a}","w":{"format":"{a}","a":"x"}}`, `"{/cat as n}{n}|{n.w}"`},
		"the field is read through the name too": {cat, `"{/cat as n}{n}|{n.w}|{n.w.a}"`},
		"only the whole is read":                 {cat, `"{/cat as n}{n}|{n}"`},
		"only the field is read":                 {cat, `"{/cat as n}{n.w}|{n.w}"`},
	} {
		t.Run(label, func(t *testing.T) {
			f, err := loadFence(t, map[string]string{"cat": c.cat, "card": c.card})
			if err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
			for i := 0; i < 40; i++ {
				if got := strings.Split(fake(t, f, "card"), "|"); got[0] != got[1] {
					t.Fatalf("draw %d = %q, want every read to print one value", i, got)
				}
			}
		})
	}
}

func TestAReadOfANameLoadsBesideAReadWhoseFormatReadsAnotherFieldOfTheSameTemplate(t *testing.T) {
	for label, c := range map[string]struct{ cat, card string }{
		"two fields of a template that isn't a constant": {
			`{"format":"{w.a.c}","w":{"format":"{a.d}","a":{"format":"{c}{d}","c":["p","q"],"d":["r","s"]}}}`,
			`"{/cat as n}{n}|{n.w}"`,
		},
	} {
		t.Run(label, func(t *testing.T) {
			if _, err := loadFence(t, map[string]string{"cat": c.cat, "card": c.card}); err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
		})
	}
}

func TestAReadOfATableRowIsRefusedBesideAReferenceItsFormatMakesToALinkedTable(t *testing.T) {
	files := nameTables()
	files["region.json"] = `{"format":"{/region.municipality.name}","rows":"region.tsv","key":"code"}`
	for label, c := range map[string]struct{ card, want string }{
		"the row and its linked table are read": {`"{/region as r}{r} {r.municipality.name}"`, "{/region.municipality.name}"},
		"a selected row and its linked table":   {`"{/region[01] as r}{r}|{r.municipality.name}"`, "{/region.municipality.name}"},
		"only the row is read":                  {`"{/region as r}{r}"`, ""},
	} {
		t.Run(label, func(t *testing.T) {
			files["card.json"] = c.card
			_, err := New(WithSeed(1), onTestDay, WithDataPath(writeFiles(t, files)))
			switch {
			case c.want == "" && err != nil:
				t.Fatalf("New() = %v, want it to load", err)
			case c.want != "" && err == nil:
				t.Fatal("New() loaded, want it refused")
			case c.want != "":
				for _, want := range []string{"{r}", c.want, "{r.municipality.name}"} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("New() = %v, want an error naming %s", err, want)
					}
				}
			}
		})
	}
}

func TestAReadLandingAboveAReferenceIsRefusedWhereRenderingItsLevelReadsTheReferenceTarget(t *testing.T) {
	const ab = `{"format":"{name}","name":["x","y"]}`
	regionRef := func(format string) map[string]string {
		files := nameTables()
		files["region.json"] = `{"format":"` + format + `","rows":"region.tsv","key":"code"}`
		return files
	}
	for label, c := range map[string]struct {
		files map[string]string
		card  string
		want  []string
	}{
		"a linked table's row, whose format reads the target": {
			regionRef("{/region.municipality.name}"),
			`"{/region as r}{r}|{r.municipality}"`,
			[]string{"{r}", "{/region.municipality.name}", "{r.municipality}"}},
		"a field, whose format reads the target": {
			map[string]string{"cat.json": `{"format":"{/cat.ab.name}","ab":` + ab + `}`},
			`"{/cat as n}{n}|{n.ab}"`,
			[]string{"{n}", "{/cat.ab.name}", "{n.ab}"}},
		"a selected row beside a reference to its column": {
			regionRef("{/region.name}"),
			`"{/region[01] as r}{r}|{r.name}"`,
			[]string{"{r}", "{/region.name}", "{r.name}"}},
		"a field whose format holds a reference of its own to the target": {
			map[string]string{"cat.json": `{"format":"{/cat.ab.name} {ab}","ab":{"format":"{name}{/cat.ab.name}","name":["x","y"]}}`},
			`"{/cat as n}{n}|{n.ab}"`,
			[]string{"{n}", "{/cat.ab.name}", "{n.ab}"}},
		"a field whose format reads a field above the target": {
			map[string]string{"cat.json": `{"format":"{/cat.ab.c}","ab":{"format":"{c.name}","c":` + ab + `}}`},
			`"{/cat as n}{n}|{n.ab}"`,
			[]string{"{n}", "{/cat.ab.c}", "{n.ab}"}},
		"a field whose format reads the target through a field": {
			map[string]string{"cat.json": `{"format":"{/cat.ab.c.name}","ab":{"format":"{c}","c":` + ab + `}}`},
			`"{/cat as n}{n}|{n.ab}"`,
			[]string{"{n}", "{/cat.ab.c.name}", "{n.ab}"}},
	} {
		t.Run(label, func(t *testing.T) {
			c.files["card.json"] = c.card
			_, err := New(WithSeed(1), onTestDay, WithDataPath(writeFiles(t, c.files)))
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

func TestAReadLandingAboveAReferenceLoadsWhereRenderingItsLevelDoesNotReadTheReferenceTarget(t *testing.T) {
	for label, c := range map[string]struct{ cat, card, want string }{
		"the level's format reads another field": {
			`{"format":"{/cat.ab.name}","ab":{"format":"{other}","name":["x","y"],"other":"k"}}`,
			`"{/cat as n}{n}|{n.ab}"`, "k"},
		"the level's format reads no field": {
			`{"format":"{/cat.ab.name}","ab":{"format":"k","name":["x","y"]}}`,
			`"{/cat as n}{n}|{n.ab}"`, "k"},
		"the level's format holds the reference itself": {
			`{"format":"{ab}","ab":{"format":"{name}{/cat.ab.name}","name":["x","y"]}}`,
			`"{/cat as n}{n}|{n.ab}"`, "="},
		"the level's format reaches the target only through a reference": {
			`{"format":"{ab}","ab":{"format":"{/cat.ab.name}","name":["x","y"]}}`,
			`"{/cat as n}{n}|{n.ab}"`, "="},
		"the level's format reads the target through a name bound to a reference": {
			`{"format":"{ab}","ab":{"format":"{/cat.ab.name as q}{q}","name":["x","y"]}}`,
			`"{/cat as n}{n}|{n.ab}"`, "="},
		"the level's format reads a name bound to another draw of the target": {
			`{"format":"{/cat.ab.name} {/cat.ab.name as q}{ab}","ab":{"format":"{q}","name":["x","y"]}}`,
			`"{/cat as n}{n}|{n.ab}"`, ""},
		"the level's format reads the target's constant": {
			`{"format":"{/cat.ab.name}","ab":{"format":"{name}","name":"k"}}`,
			`"{/cat as n}{n}|{n.ab}"`, "k"},
	} {
		t.Run(label, func(t *testing.T) {
			f, err := loadFence(t, map[string]string{"cat": c.cat, "card": c.card})
			if err != nil {
				t.Fatalf("New() = %v, want it to load", err)
			}
			got := strings.SplitN(fake(t, f, "card"), "|", 2)
			if c.want == "=" && got[0] != got[1] || c.want != "" && c.want != "=" && got[1] != c.want {
				t.Errorf("card = %q, want it to end %q, or both reads to agree where that is =", got, c.want)
			}
		})
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
		"a reference back to a level above the name's start": {map[string]string{
			"cat":  `{"format":"{x}","x":{"format":"{z}","y":"{/cat.x}","z":["a","b"]}}`,
			"card": `"{/cat.x.y as n}{n}{n}"`,
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
		"a field bound to a name, beside a sibling that draws": {
			`{"format":"{a.b as r}{r}","a":{"format":"{b}","b":["x","x"],"c":["p","q"]}}`,
			`"{/cat as n}{n}|{n.a.c}"`, "x"},
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
		"two fields of a plain template, read whole beside them": {
			`{"format":"{w.a} {w.b}","w":{"format":"{a}","a":["x","y"],"b":["p"]}}`,
			`"{/cat as n}{n}|{n.w}"`, `renders field "w" twice`},
		"a plain template whose leaf draws, bound to a name": {
			`{"format":"{a.b as r}{r}","a":{"format":"{b}","b":{"format":"{q}","q":["x","y"]}}}`,
			`"{/cat as n}{n}|{n.a}"`, `through name "r"`},
		"a field read twice inside a choice's item": {
			`{"format":"{w}","w":[{"format":"{a}-{a}","a":["x","y"]},{"format":"{a}-{a}","a":["x","y"]}]}`,
			`"{/cat as n}{n}|{n.w.a}"`, `renders field "a" twice`},
		"a field read twice, where a reference draws it afresh too": {
			`{"format":"{w}-{w} {/cat.w}","w":["x","y"]}`,
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

func TestATableFormatReadingAColumnTwiceLoadsBesideAReadOfThatColumn(t *testing.T) {
	f, err := New(WithSeed(1), onTestDay, WithDataPath(writeFiles(t, map[string]string{
		"row.json": `{"format":"{name}-{name}","rows":"row.tsv","key":"code","name":"name"}`,
		"row.tsv":  "code\tname\nA\tx\nB\ty\n",
		"co.json":  `"{/row as r}{r} {r.name}"`,
	})))
	if err != nil {
		t.Fatalf("New() = %v, want it loaded", err)
	}
	for range 20 {
		got := fake(t, f, "co")
		if got != "x-x x" && got != "y-y y" {
			t.Fatalf("co = %q, want one row's name throughout", got)
		}
	}
}
