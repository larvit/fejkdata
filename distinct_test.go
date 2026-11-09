package fejkdata

import (
	"runtime"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDistinctDrawsDistinctItems(t *testing.T) {
	f := engine(1)
	for src, want := range map[string][]string{
		`{"format":"{w distinct 2 joined ','}","w":["a","b","c"]}`:                        {"a,b", "a,c", "b,a", "b,c", "c,a", "c,b"},
		`{"format":"{w distinct 3 joined ''}","w":["a","b","c"]}`:                         {"abc", "acb", "bac", "bca", "cab", "cba"},
		`{"format":"{w distinct 2 joined ','}","w":["a","a","b"]}`:                        {"a,a", "a,b", "b,a"},
		`{"format":"{w distinct 2 joined ','}","w":[{"format":"{v}","v":["x","y"]},"z"]}`: {"x,y", "y,x", "x,z", "z,x", "y,z", "z,y"},
		`{"format":"{w where length 3 distinct 2 joined ' '}","w":["cat","dog","horse"]}`: {"cat dog", "dog cat"},
		`{"format":"{w distinct 2 joined ','}","w":["a","b",{"format":"c","weight":0}]}`:  {"a,b", "b,a"},
	} {
		t.Run(src, func(t *testing.T) {
			drawsExactly(t, src, func() string { return mustRender(t, f, src) }, want...)
		})
	}
}

func TestDistinctBindsMembersCountedFromOne(t *testing.T) {
	f := engine(1)
	for src, want := range map[string][]string{
		`{"format":"{w distinct 3 as e}{e[1]}{e[2]}{e[3]}","w":["a","b","c"]}`:           {"abc", "acb", "bac", "bca", "cab", "cba"},
		`{"format":"{w distinct 2 as e}{e[1]}{e[2]}{e[1]}","w":["a","b"]}`:               {"aba", "bab"},
		`{"format":"{w distinct 2 joined '-' as e}{e}/{e[2]}","w":["a","b"]}`:            {"a-b/b", "b-a/a"},
		`{"format":"{w distinct 2 as e}{uppercase(e[1])}{e[1]}","w":["a","b","c"]}`:      {"Aa", "Bb", "Cc"},
		`{"format":"{w where length 1 distinct 2 as e}{e[2]}{e[1]}","w":["a","b","cc"]}`: {"ab", "ba"},
	} {
		t.Run(src, func(t *testing.T) {
			drawsExactly(t, src, func() string { return mustRender(t, f, src) }, want...)
		})
	}
}

func TestDistinctMemberIsAnOperand(t *testing.T) {
	f := engine(1)
	for src, want := range map[string][]string{
		`{"format":"{w distinct 2 as m}{calc(m[1]*2)},{m[1]}","w":["3","4"]}`:                           {"6,3", "8,4"},
		`{"format":"{w distinct 2 as m}{m[2]}:{v where length m[2]}","w":["3","4"],"v":["cat","bird"]}`: {"3:cat", "4:bird"},
	} {
		t.Run(src, func(t *testing.T) {
			drawsExactly(t, src, func() string { return mustRender(t, f, src) }, want...)
		})
	}
}

func TestDistinctFillsARecordsColumnsFromOnePick(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"x": `"x"`}))
	rt, err := f.NewRecordTemplate(`{"format":"{crew distinct 4 as e}","let":{"crew":["ann","bo","cy","di","ed"]},"e1":"{e[1]}","e2":"{e[2]}","e3":"{e[3]}","e4":"{e[4]}"}`)
	if err != nil {
		t.Fatalf("NewRecordTemplate: %v", err)
	}
	eachDraw(200, func() string { return strings.Join(columnValues(rt.Fake()), "|") }, func(joined string) {
		got := strings.Split(joined, "|")
		sorted := slices.Clone(got)
		slices.Sort(sorted)
		if len(got) != 4 || len(slices.Compact(sorted)) != 4 {
			t.Fatalf("columns = %q, want four crew members, none twice", got)
		}
	})
}

func columnValues(r *Record) []string {
	var values []string
	for _, c := range r.Columns() {
		values = append(values, c.Value)
	}
	return values
}

func TestDistinctDrawsDistinctRowsWhateverTheyPrint(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"animal.json": `{"format":"{name}","rows":"animal.tsv","key":"code"}`,
		"animal.tsv":  "code\tname\nC\tcat\nH\thorse\nX\tcat\n",
	}))
	for src, want := range map[string][]string{
		`"{/animal.name distinct 2 joined ','}"`: {"cat,horse", "cat,cat", "horse,cat"},
		`"{/animal.code distinct 3 joined ''}"`:  {"CHX", "CXH", "HCX", "HXC", "XCH", "XHC"},
		`"{/animal distinct 3 as a}{a[3]}"`:      {"cat", "horse"},
	} {
		t.Run(src, func(t *testing.T) {
			tp, err := f.NewTemplate(src)
			if err != nil {
				t.Fatalf("NewTemplate(%s): %v", src, err)
			}
			drawsExactly(t, src, tp.Fake, want...)
		})
	}
}

// The first draw takes a by its weight of 6 in 10; the second weighs b against c as 3 to 1, so a,b
// is 0.6 × 3/4 = 0.45. A first b leaves a against c as 6 to 1, so b,a is 0.3 × 6/7 ≈ 0.257.
func TestDistinctWeighsEachMemberAmongTheItemsLeft(t *testing.T) {
	f := engine(3)
	src := `{"format":"{w distinct 2 joined ','}","w":[{"format":"a","weight":6},{"format":"b","weight":3},{"format":"c","weight":1}]}`
	seen := tallyDraws(6000, func() string { return mustRender(t, f, src) })
	if n := seen["a,b"]; n < 2550 || n > 2850 {
		t.Errorf("a,b drew %d of 6000, want about 2700: %v", n, seen)
	}
	if n := seen["b,a"]; n < 1410 || n > 1680 {
		t.Errorf("b,a drew %d of 6000, want about 1543: %v", n, seen)
	}
}

func TestDistinctRefusesAtLoad(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"x": `"x"`}))
	for src, want := range map[string]string{
		`{"format":"{w distinct 4 joined ','}","w":["a","b","c"]}`:                                                                             "renders 3 of weight above 0",
		`{"format":"{w distinct 3 joined ','}","w":["a","b",{"format":"c","weight":0}]}`:                                                       "renders 2 of weight above 0",
		`{"format":"{w where length 3 distinct 3 joined ','}","w":["cat","dog","horse"]}`:                                                      "only 2 pass",
		`{"format":"{n as k}{w where length k distinct 2 joined ','}","n":["3","4"],"w":["cat","dog","bird"]}`:                                 "where k is 4",
		`{"format":"{w distinct 2 as e}{e}","w":["a","b"]}`:                                                                                    "{e[1]}",
		`{"format":"{w distinct 2 as e}{e[3]}","w":["a","b","c"]}`:                                                                             "{e[1]} to {e[2]}",
		`{"format":"{w distinct 2 as e}{e[0]}","w":["a","b","c"]}`:                                                                             "{e[1]} to {e[2]}",
		`{"format":"{w distinct 2 as e}{e[x]}","w":["a","b","c"]}`:                                                                             "{e[1]} to {e[2]}",
		`{"format":"{w distinct 2 as e}{e[1].x}","w":["a","b"]}`:                                                                               "read whole",
		`{"format":"{w distinct 2 as e}{e.x}","w":["a","b"]}`:                                                                                  "read whole",
		`{"format":"{w as e}{e[1]}","w":["a","b"]}`:                                                                                            "select a row",
		`{"format":"{w distinct 2 joined ','}","w":["{lower(5)}","b","c"]}`:                                                                    "fixed at load",
		`{"format":"{w distinct 1 joined ''}","w":[{"format":"{v distinct 1 joined ''}","v":["a","bb"]}]}`:                                     "a clause of its own",
		`{"format":"{w distinct 2 as  e}{e[1]}","w":["a","b"]}`:                                                                                "write {w distinct 2 as e}",
		`{"format":"{w as x}{x distinct 2 joined ','}","w":["a","b"]}`:                                                                         "drawn once",
		`{"format":"{w distinct 2 joined ',' as e}{v where length e}","w":["3","4","5"],"v":["cat","bird","horse"]}`:                           "joins several values",
		`{"format":"{w distinct 2 joined ',' as e}{calc(e*2)}","w":["3","4","5"]}`:                                                             "joins several values",
		`{"format":"{d as s}{w distinct 2 joined ',' as e}{adddays(s,e)}","d":"{date(2026-01-01,2026-12-31,'2006-01-02')}","w":["3","4","5"]}`: "joins several values",
	} {
		t.Run(src, func(t *testing.T) {
			if _, err := f.NewTemplate(src); err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("NewTemplate(%s) = %v, want an error holding %q", src, err, want)
			}
		})
	}
}

func TestDistinctRefusesInATable(t *testing.T) {
	_, err := New(WithSeed(1), onTestDay, WithDataPath(writeFiles(t, map[string]string{
		"pet.json": `{"format":"{name distinct 2 joined ','}","rows":"pet.tsv"}`,
		"pet.tsv":  "name\ncat\ndog\n",
	})))
	if err == nil || !strings.Contains(err.Error(), "in a template reading the table") {
		t.Errorf("New = %v, want a table's format drawing distinct values refused, naming where to draw them", err)
	}
}

func TestDistinctJoinedColumnIsTextAndAMemberTheColumnItReads(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{
		"src": `{"format":"","n":[{"format":"10","datatype":"integer"},{"format":"20","datatype":"integer"},{"format":"30","datatype":"integer"}]}`,
		"row": `{"format":"{/src.n distinct 2 joined ',' as e}","pair":"{/src.n distinct 2 joined ','}","both":"{e}","first":"{e[1]}","own":"{/src.n distinct 2 as y}{y[1]}","kept":"{/src.n where length 2 as z}{z}","lit":{"format":"{/src.n where length 2 as q}5","datatype":"integer"}}`,
	}))
	r, err := f.FakeRecord("row")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range r.Columns() {
		t.Run(c.Name, func(t *testing.T) {
			if want := distinctColumnType(c.Name); c.DataType != want {
				t.Errorf("column %s = %q of %s, want %s", c.Name, c.Value, c.DataType, want)
			}
		})
	}
}

func distinctColumnType(name string) DataType {
	if name == "first" || name == "own" || name == "kept" || name == "lit" {
		return DataTypeInteger
	}
	return DataTypeString
}

func TestDistinctJoinedIsNoTypedValue(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"n": `["1","2","3"]`}))
	for _, src := range []string{
		`{"format":"","c":{"format":"{/n distinct 2 joined ','}","datatype":"integer"}}`,
		`{"format":"{w distinct 2 joined ',' as e}","c":{"format":"{e}","datatype":"integer"},"w":["1","2","3"]}`,
	} {
		t.Run(src, func(t *testing.T) {
			if _, err := f.NewRecordTemplate(src); err == nil || !strings.Contains(err.Error(), "prints text, not an integer") {
				t.Errorf("NewRecordTemplate(%s) = %v, want an integer column of joined values refused", src, err)
			}
		})
	}
}

func TestDistinctDrawsTheShippedWords(t *testing.T) {
	t.Parallel()
	f := shipped(t)
	eachDraw(50, func() string {
		return tmpl(t, f, "{/en_US.word where length 5 distinct 3 as w}{w[1]} {w[2]} {w[3]}")
	}, func(got string) {
		words := strings.Fields(got)
		if len(words) != 3 || slices.ContainsFunc(words, func(w string) bool { return utf8.RuneCountInString(w) != 5 }) {
			t.Fatalf("words = %q, want three words of five letters", words)
		}
	})
}

func TestDistinctDrawsTheShippedWeekdays(t *testing.T) {
	t.Parallel()
	f := shipped(t)
	eachDraw(50, func() string {
		return tmpl(t, f, "{/sv_SE.weekday.code distinct 2 joined ','}")
	}, func(got string) {
		days := strings.Split(got, ",")
		if len(days) != 2 || days[0] == days[1] || slices.ContainsFunc(days, func(d string) bool { return len(d) != 1 || d < "1" || d > "7" }) {
			t.Fatalf("days = %q, want two different weekday codes", days)
		}
	})
}

func TestDistinctDrawCostsItsCountNotTheTable(t *testing.T) {
	f := shipped(t)
	tp, err := f.NewTemplate(`"{/en_US.word distinct 3 joined ','}"`)
	if err != nil {
		t.Fatalf("NewTemplate: %v", err)
	}
	tp.Fake()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	eachDraw(100, tp.Fake, func(string) {})
	runtime.ReadMemStats(&after)
	if per := (after.TotalAlloc - before.TotalAlloc) / 100; per > 4096 {
		t.Errorf("a draw of 3 distinct words allocates %d bytes, want 4096 or fewer whatever the table's size", per)
	}
}
