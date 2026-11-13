package fejkdata

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type structPlace struct {
	City string `fake:"{/place as p}{p.city}"`
	Zip  string `fake:"{p.zip}"`
}

type structUser struct {
	Active bool   `fake:"[\"true\",\"false\"]"`
	Age    uint8  `fake:"{int(18,99)}"`
	Email  string `fake:"{lowercase(p.first)}@example.com"`
	First  string `fake:"{/person as p}{p.first}"`
	Home   structPlace
	ID     int64   `fake:"{seq()}"`
	Last   string  `fake:"{p.last}"`
	Level  uint8   `fake:"{float(0,255,0)}"`
	Nick   *string `fake:"[null,\"bo\"]"`
	Note   string
	Rank   *int         `fake:"{\"format\":\"{r}\",\"r\":[\"1\",\"2\"]}"`
	Score  float32      `fake:"{float(0,1,2)}"`
	Skip   *structPlace `fake:"-"`
	Work   *structPlace
	hidden structPlace
}

type structGiven struct {
	First string `fake:"{/person as p}{p.first}"`
}

type StructFamily struct {
	Last string `fake:"person.last"`
}

type structEmployee struct {
	structGiven
	*StructFamily
	Email string `fake:"{lowercase(p.first)}@example.com"`
}

type structLink struct {
	Name string `fake:"person.first"`
	Next *structLink
}

func structData(t *testing.T) *Generator {
	t.Helper()
	return newGenerator(t, writeData(t, map[string]string{
		"person": `[{"format":"{first} {last}","first":"Ada","last":"Lovelace"},{"format":"{first} {last}","first":"Bo","last":"Ek","born":"1815"}]`,
		"place":  `[{"format":"{city}","city":"Stockholm","zip":"111 22"},{"format":"{city}","city":"Tranås","zip":"573 31","region":"F"}]`,
		"mid":    `{"format":"","score":["{/src.score}",{"format":"5","datatype":"integer"}]}`,
		"src":    `{"format":"","code":[null,"200","404"],"del":null,"score":[null,{"format":"{int(1,9)}","datatype":"integer"}]}`,
		"trip":   `{"format":"","leg":[{"format":"{to}","to":"Oslo"},{"format":"{to}","to":"Rome"}]}`,
	}), WithSeed(1), onTestDay)
}

type structOfColumns struct {
	Code  *string `fake:"src.code"`
	Codes *string `fake:"[\"{/src.score}\",\"{/src.code}\"]"`
	Del   *int64  `fake:"src.del"`
	Label string  `fake:"n={/src.score}"`
	Mixed *string `fake:"[\"{/src.score}\",\"x\"]"`
	Pair  *int64  `fake:"[\"{/src.score}\",\"5\"]"`
	Score *int64  `fake:"src.score"`
}

func pointedText(p any) string {
	switch p := p.(type) {
	case *string:
		if p != nil {
			return *p
		}
	case *int64:
		if p != nil {
			return strconv.FormatInt(*p, 10)
		}
	}
	return "nil"
}

func eachStructFilled[T any](t *testing.T, f *Generator, n int, check func(v T)) {
	t.Helper()
	for i := 0; i < n; i++ {
		var v T
		if err := f.FakeStruct(&v); err != nil {
			t.Fatal(err)
		}
		check(v)
	}
}

func someStructFilled[T any](t *testing.T, f *Generator, n int, found func(v T) bool) bool {
	t.Helper()
	for i := 0; i < n; i++ {
		var v T
		if err := f.FakeStruct(&v); err != nil {
			t.Fatal(err)
		}
		if found(v) {
			return true
		}
	}
	return false
}

func TestFakeStructFieldTaggedWithAColumnIsThatColumn(t *testing.T) {
	eachStructFilled(t, structData(t), 200, func(v structOfColumns) {
		score, code := pointedText(v.Score), pointedText(v.Code)
		switch {
		case pointedText(v.Del) != "nil":
			t.Fatalf("Del = %s, want nil from a column only ever null", pointedText(v.Del))
		case !digitOrNil(pointedText(v.Mixed)) && pointedText(v.Mixed) != "x", !digitOrNil(pointedText(v.Pair)) && pointedText(v.Pair) != "5", !digitOrNil(pointedText(v.Codes)) && pointedText(v.Codes) != "200" && pointedText(v.Codes) != "404":
			t.Fatalf("Mixed %s, Pair %s, Codes %s: want each the literal or a draw of the column it reads", pointedText(v.Mixed), pointedText(v.Pair), pointedText(v.Codes))
		case score != "nil" && !digitOrNil(score), code != "nil" && code != "200" && code != "404":
			t.Fatalf("Score %s, Code %s: want each nil or a draw of its column", score, code)
		}
		if v.Code != nil && *v.Code != "200" && *v.Code != "404" {
			t.Fatalf("Code = %q, want nil or a code, never a null's \"\"", *v.Code)
		}
		if label, _ := strings.CutPrefix(v.Label, "n="); !strings.HasPrefix(v.Label, "n=") || label != "" && !digitOrNil(label) {
			t.Fatalf("Label %q, want n= and a draw of src.score, a null as \"\"", v.Label)
		}
	})
}

func TestFakeStructPointerTaggedWithANullableColumnIsNilOnlySometimes(t *testing.T) {
	nils := map[string]int{}
	eachStructFilled(t, structData(t), 200, func(v structOfColumns) {
		if v.Code == nil {
			nils["code"]++
		}
		if v.Score == nil {
			nils["score"]++
		}
	})
	for _, name := range []string{"code", "score"} {
		t.Run(name, func(t *testing.T) {
			if n := nils[name]; n == 0 || n == 200 {
				t.Errorf("%s was nil %d times in 200 fills, want both outcomes", name, n)
			}
		})
	}
}

type structOfPathsAndNames struct {
	A    string `fake:"trip.leg"`
	B    string `fake:"trip.leg"`
	Leg  string `fake:"{/trip as r}{r.leg}"`
	To   string `fake:"{r.leg.to}"`
	Name string `fake:"{/person as p}{p}"`
	Last string `fake:"{p.last}"`
}

func TestFakeStructNameKeepsOnePick(t *testing.T) {
	eachStructFilled(t, structData(t), 100, func(v structOfPathsAndNames) {
		if v.Leg != v.To || v.Name != "Ada "+v.Last && v.Name != "Bo "+v.Last {
			t.Fatalf("%+v, want the reads of r one leg, and of p one person", v)
		}
	})
}

func TestFakeStructPathTagsDrawAfresh(t *testing.T) {
	apart := false
	eachStructFilled(t, structData(t), 100, func(v structOfPathsAndNames) {
		apart = apart || v.A != v.B
	})
	if !apart {
		t.Error("two tags of trip.leg drew one leg in 100 fills, want each path tag a draw of its own")
	}
}

func eachStructUserAndItsTwin(t *testing.T, check func(i int, u, twin structUser, oldNick string)) {
	t.Helper()
	a, b := structData(t), structData(t)
	for i := 0; i < 100; i++ {
		oldNick, oldTwinNick := "old", "old"
		u, twin := structUser{Nick: &oldNick, Note: "keep"}, structUser{Nick: &oldTwinNick, Note: "keep"}
		if err := a.FakeStruct(&u); err != nil {
			t.Fatal(err)
		}
		if err := b.FakeStruct(&twin); err != nil {
			t.Fatal(err)
		}
		check(i, u, twin, oldNick)
	}
}

func TestFakeStructFillsTaggedFields(t *testing.T) {
	people := map[string]string{"Ada": "Lovelace", "Bo": "Ek"}
	zips := map[string]string{"Stockholm": "111 22", "Tranås": "573 31"}
	eachStructUserAndItsTwin(t, func(i int, u, twin structUser, oldNick string) {
		switch {
		case !reflect.DeepEqual(u, twin):
			t.Fatalf("same seed diverged: %+v != %+v", u, twin)
		case people[u.First] != u.Last || u.Email != strings.ToLower(u.First)+"@example.com":
			t.Fatalf("person fields %q %q %q, want one person read through p across the struct", u.First, u.Last, u.Email)
		case zips[u.Home.City] != u.Home.Zip || u.Work == nil || zips[u.Work.City] != u.Work.Zip:
			t.Fatalf("places %+v, %+v, want each nested struct one place, the pointer allocated", u.Home, u.Work)
		case u.ID != int64(i+1) || u.Age < 18 || u.Age > 99 || u.Score < 0 || u.Score > 1 || u.Rank == nil || (*u.Rank != 1 && *u.Rank != 2):
			t.Fatalf("typed fields %+v, want each the value its tag draws", u)
		case u.Nick != nil && *u.Nick != "bo", u.Note != "keep", u.hidden != (structPlace{}), u.Skip != nil, oldNick != "old":
			t.Fatalf("%+v: want Nick nil or bo, and the untagged fields left as they were", u)
		}
	})
}

func TestFakeStructDrawsBothOutcomesOfABooleanAndOfANullableField(t *testing.T) {
	actives, nils := 0, 0
	eachStructUserAndItsTwin(t, func(_ int, u, _ structUser, _ string) {
		if u.Active {
			actives++
		}
		if u.Nick == nil {
			nils++
		}
	})
	if actives == 0 || actives == 100 || nils == 0 || nils == 100 {
		t.Errorf("100 draws gave %d active and %d nil nicks, want both outcomes of each", actives, nils)
	}
}

func TestFakeStructFillsEmbeddedFieldsIntoItsRecord(t *testing.T) {
	eachStructFilled(t, structData(t), 100, func(e structEmployee) {
		if e.StructFamily == nil || e.Last != "Lovelace" && e.Last != "Ek" || e.Email != strings.ToLower(e.First)+"@example.com" {
			t.Fatalf("%+v, %+v: want a promoted field's name read by the struct's own, the embedded pointer allocated", e, e.StructFamily)
		}
	})
}

func TestFakeStructLeavesAnEmbeddedStructTaggedDashUnfilled(t *testing.T) {
	f := structData(t)
	var skipped struct {
		structGiven `fake:"-"`
		Email       string `fake:"{lowercase(/person.first)}@example.com"`
	}
	if err := f.FakeStruct(&skipped); err != nil || skipped.First != "" || skipped.Email == "" {
		t.Errorf("FakeStruct = %v, %+v; want the embedded struct under fake:\"-\" left unfilled beside the filled field", err, skipped)
	}
}

func TestFakeStructDrawsANestedStructApart(t *testing.T) {
	apart := someStructFilled(t, structData(t), 100, func(trip struct{ From, To structPlace }) bool {
		return trip.From.City != trip.To.City
	})
	if !apart {
		t.Error("From and To drew one place in 100 trips; a nested struct is a record of its own, so each draws apart")
	}
}

func TestFakeStructLeavesAPointerBackAlone(t *testing.T) {
	var l structLink
	if err := structData(t).FakeStruct(&l); err != nil || l.Name == "" || l.Next != nil {
		t.Errorf("FakeStruct = %v, %+v; want Name filled and Next, a pointer back to the struct being filled, left nil", err, l)
	}
}

type structOfPathSpellings struct {
	Path   string `fake:"person.first"`
	Ref    string `fake:"{/person.first}"`
	Quoted string `fake:"\"{/person.first}\""`
	Slash  string `fake:"/person.first"`
	Kept   string `fake:"-"`
}

func filledStructOfPathSpellings(t *testing.T) structOfPathSpellings {
	t.Helper()
	v := structOfPathSpellings{Kept: "kept"}
	if err := structData(t).FakeStruct(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestFakeStructReadsAPathInEachSpelling(t *testing.T) {
	v := filledStructOfPathSpellings(t)
	for name, got := range map[string]string{"Path": v.Path, "Ref": v.Ref, "Quoted": v.Quoted, "Slash": v.Slash} {
		t.Run(name, func(t *testing.T) {
			if got != "Ada" && got != "Bo" {
				t.Errorf("%s = %q, want a first name", name, got)
			}
		})
	}
}

func TestFakeStructLeavesAFieldTaggedDashAsItWas(t *testing.T) {
	v := filledStructOfPathSpellings(t)
	if v.Kept != "kept" {
		t.Errorf(`Kept = %q, want fake:"-" to leave it as it was`, v.Kept)
	}
}

func TestFakeStructTakesTheDatatypeItsGoTypeSets(t *testing.T) {
	v := struct {
		N int    `fake:"{\"format\":\"{int(1,9)}\",\"datatype\":\"integer\"}"`
		S string `fake:"{\"format\":\"x\",\"datatype\":\"string\"}"`
	}{}
	if err := structData(t).FakeStruct(&v); err != nil || v.N < 1 || v.N > 9 || v.S != "x" {
		t.Errorf("FakeStruct = %+v, %v, want each tag's datatype taken as its Go type's", v, err)
	}
}

type structOfBareValues struct {
	N int    `fake:"[5, 6]"`
	B bool   `fake:"[true, false]"`
	S string `fake:"[7.50]"`
}

func TestFakeStructTypesABareNumberOrBooleanByItsGoType(t *testing.T) {
	seen := map[bool]bool{}
	eachStructFilled(t, structData(t), 30, func(v structOfBareValues) {
		if (v.N != 5 && v.N != 6) || v.S != "7.50" {
			t.Fatalf("FakeStruct = %+v, want each value typed by its field", v)
		}
		seen[v.B] = true
	})
	if len(seen) != 2 {
		t.Errorf("B took %v in 30 fills, want true and false", seen)
	}
}

func TestFakeStructWithNoTagsFillsNothing(t *testing.T) {
	v := struct{ A string }{A: "kept"}
	if err := structData(t).FakeStruct(&v); err != nil || v.A != "kept" {
		t.Errorf("FakeStruct = %+v, %v, want nothing filled", v, err)
	}
}

func TestFakeStructErrors(t *testing.T) {
	f := structData(t)
	for _, c := range []struct {
		name string
		v    any
		want string
	}{
		{"a struct, not a pointer", structUser{}, "fills a struct through a non-nil pointer, got fejkdata.structUser"},
		{"a nil pointer to a struct", (*structUser)(nil), "through a non-nil pointer"},
		{"a pointer to an int", new(int), "through a non-nil pointer"},
		{"nil", nil, "through a non-nil pointer"},
		{"an unexported field", &struct {
			a string `fake:"person.first"`
		}{}, ".a: unexported"},
		{"a slice field", &struct {
			A []string `fake:"person.first"`
		}{}, ".A: a fake tag fills a string, bool, integer or float field, or a pointer to one, not []string"},
		{"a pointer to a pointer", &struct {
			A **int `fake:"{int(1,9)}"`
		}{}, "not **int"},
		{"a tagged struct field", &struct {
			A structPlace `fake:"place"`
		}{}, ".A: a struct field fills from the tags on its own fields"},
		{"a datatype other than the Go type's on an int", &struct {
			A int `fake:"{\"format\":\"{int(1,9)}\",\"datatype\":\"number\"}"`
		}{}, `its Go type int sets the datatype integer; drop "datatype"`},
		{"a datatype other than the Go type's on a string", &struct {
			A string `fake:"{\"format\":\"1\",\"datatype\":\"integer\"}"`
		}{}, `its Go type string sets the datatype string; drop "datatype"`},
		{"a null item in an int", &struct {
			A int `fake:"[null,\"{int(1,9)}\"]"`
		}{}, "can draw null, which int cannot hold; make it *int"},
		{"a nullable integer column in an int64", &struct {
			A int64 `fake:"src.score"`
		}{}, "can draw null, which int64 cannot hold; make it *int64"},
		{"a nullable string column in a string", &struct {
			A string `fake:"src.code"`
		}{}, "can draw null, which string cannot hold; make it *string"},
		{"an integer column in a *bool", &struct {
			A *bool `fake:"src.score"`
		}{}, ".A (*bool): {int(1,9)} prints an integer, not a boolean"},
		{"a column reading a nullable one in an int64", &struct {
			A int64 `fake:"mid.score"`
		}{}, "can draw null, which int64 cannot hold; make it *int64"},
		{"text digits in an int", &struct {
			A int `fake:"{digits(3)}"`
		}{}, ".A (int): {digits(3)} prints text, not an integer"},
		{"a name in an int", &struct {
			A int `fake:"person.first"`
		}{}, `"Ada" is not an integer`},
		{"an operand that prints nothing", &struct {
			A int `fake:"{\"format\":\"{calc(d + 1)}\",\"d\":\"{digits(0)}\"}"`
		}{}, "{digits(0)} prints nothing, which is no number"},
		{"an int8 past its range", &struct {
			A int8 `fake:"{int(0,300)}"`
		}{}, `"{int(0,300)}" is not proven within int8; narrow it to that range, or make the field int64`},
		{"a uint below zero", &struct {
			A uint `fake:"{int(-1,5)}"`
		}{}, `"{int(-1,5)}" is not proven within uint; narrow it to that range, or make the field int64`},
		{"a float32 past its range", &struct {
			A float32 `fake:"[\"1\",\"1e39\"]"`
		}{}, `"1e39" is not proven within float32; narrow it to that range, or make the field float64`},
		{"a counter in an int32", &struct {
			A int32 `fake:"{seq()}"`
		}{}, `"{seq()}" is not proven within int32; narrow it to that range, or make the field int64`},
		{"an integer in a bool", &struct {
			A bool `fake:"{int(0,1)}"`
		}{}, "prints an integer, not a boolean"},
		{"a path to no entry", &struct {
			A string `fake:"nope.x"`
		}{}, `no entry "nope"`},
		{"an alternation without braces", &struct {
			A string `fake:"a|b"`
		}{}, `contains "|"`},
		{"an empty tag", &struct {
			A string `fake:""`
		}{}, "is empty"},
		{"a selector with nothing before it", &struct {
			A string `fake:"[abc]"`
		}{}, `"["`},
		{"a relative reference", &struct {
			A string `fake:"{.person.first} x"`
		}{}, "write {/person.first}"},
		{"an unexported embedded pointer", &struct{ *structGiven }{}, "an unexported embedded pointer field cannot be set"},
		{"a field hiding an embedded one", &struct {
			structGiven
			First string `fake:"person.last"`
		}{}, "struct.structGiven.First: hidden by another field named First"},
		{"a read of no field", &struct {
			A string `fake:"{x}"`
		}{}, `no field "x"`},
		{"a nested struct's field", &struct {
			Trip struct {
				A int `fake:"{digits(3)}"`
			}
		}{}, ": struct.Trip.A (int): {digits(3)} prints text"},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := f.FakeStruct(c.v)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("FakeStruct(%T) = %v, want an error containing %q", c.v, err, c.want)
				return
			}
			if again := f.FakeStruct(c.v); again == nil || again.Error() != err.Error() {
				t.Errorf("FakeStruct(%T) again = %v, want the first call's error, %v", c.v, again, err)
			}
		})
	}
}

func binaryTreeOfStructs(leaf reflect.Type) reflect.Type {
	tree := leaf
	for depth := 1; depth <= 8; depth++ {
		tree = reflect.StructOf([]reflect.StructField{{Name: "L", Type: reflect.PointerTo(tree)}, {Name: "R", Type: reflect.PointerTo(tree)}})
	}
	return tree
}

func TestFakeStructBoundsTheStructsATypeReaches(t *testing.T) {
	f := structData(t)
	place := reflect.TypeOf(structPlace{})
	tree := binaryTreeOfStructs(place)
	fields := []reflect.StructField{{Name: "L", Type: reflect.PointerTo(tree)}, {Name: "R", Type: reflect.PointerTo(tree)}, {Name: "P", Type: place}}
	if err := f.FakeStruct(reflect.New(reflect.StructOf(fields)).Interface()); err != nil {
		t.Fatalf("a type reaching 1024 structs: FakeStruct = %v, want it filled", err)
	}
	embedded := reflect.StructField{Name: "StructFamily", Type: reflect.TypeOf(StructFamily{}), Anonymous: true}
	err := f.FakeStruct(reflect.New(reflect.StructOf(append(fields, embedded))).Interface())
	if err == nil || !strings.Contains(err.Error(), "more than 1024 structs") || !strings.Contains(err.Error(), `leave a struct field unfilled with fake:"-"`) {
		t.Errorf("a type reaching 1025 structs, the last embedded: FakeStruct = %v, want it refused naming the cap and fake:\"-\"", err)
	}
}

func digitOrNil(s string) bool {
	return s == "nil" || len(s) == 1 && s >= "1" && s <= "9"
}
