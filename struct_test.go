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
	}), WithSeed(1))
}

func TestFakeStructFieldTaggedWithAColumnIsThatColumn(t *testing.T) {
	f := structData(t)
	nils := map[string]int{}
	show := func(p any) string {
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
	for i := 0; i < 200; i++ {
		var v struct {
			Code  *string `fake:"src.code"`
			Codes *string `fake:"[\"{/src.score}\",\"{/src.code}\"]"`
			Del   *int64  `fake:"src.del"`
			Label string  `fake:"n={/src.score}"`
			Mixed *string `fake:"[\"{/src.score}\",\"x\"]"`
			Pair  *int64  `fake:"[\"{/src.score}\",\"5\"]"`
			Score *int64  `fake:"src.score"`
		}
		if err := f.FakeStruct(&v); err != nil {
			t.Fatal(err)
		}
		score, code := show(v.Score), show(v.Code)
		switch {
		case show(v.Del) != "nil":
			t.Fatalf("Del = %s, want nil from a column only ever null", show(v.Del))
		case !digitOrNil(show(v.Mixed)) && show(v.Mixed) != "x", !digitOrNil(show(v.Pair)) && show(v.Pair) != "5", !digitOrNil(show(v.Codes)) && show(v.Codes) != "200" && show(v.Codes) != "404":
			t.Fatalf("Mixed %s, Pair %s, Codes %s: want each the literal or a draw of the column it reads", show(v.Mixed), show(v.Pair), show(v.Codes))
		case score != "nil" && !digitOrNil(score), code != "nil" && code != "200" && code != "404":
			t.Fatalf("Score %s, Code %s: want each nil or a draw of its column", score, code)
		}
		switch {
		case v.Code == nil:
			nils["code"]++
		case *v.Code != "200" && *v.Code != "404":
			t.Fatalf("Code = %q, want nil or a code, never a null's \"\"", *v.Code)
		}
		if v.Score == nil {
			nils["score"]++
		}
		if label, _ := strings.CutPrefix(v.Label, "n="); !strings.HasPrefix(v.Label, "n=") || label != "" && !digitOrNil(label) {
			t.Fatalf("Label %q, want n= and a draw of src.score, a null as \"\"", v.Label)
		}
	}
	for _, name := range []string{"code", "score"} {
		if n := nils[name]; n == 0 || n == 200 {
			t.Errorf("%s was nil %d times in 200 fills, want both outcomes", name, n)
		}
	}
}

func TestFakeStructPathTagsDrawAfreshAndANameKeepsOnePick(t *testing.T) {
	f := structData(t)
	var v struct {
		A    string `fake:"trip.leg"`
		B    string `fake:"trip.leg"`
		Leg  string `fake:"{/trip as r}{r.leg}"`
		To   string `fake:"{r.leg.to}"`
		Name string `fake:"{/person as p}{p}"`
		Last string `fake:"{p.last}"`
	}
	apart := false
	for i := 0; i < 100; i++ {
		if err := f.FakeStruct(&v); err != nil {
			t.Fatal(err)
		}
		if v.Leg != v.To || v.Name != "Ada "+v.Last && v.Name != "Bo "+v.Last {
			t.Fatalf("%+v, want the reads of r one leg, and of p one person", v)
		}
		apart = apart || v.A != v.B
	}
	if !apart {
		t.Error("two tags of trip.leg drew one leg in 100 fills, want each path tag a draw of its own")
	}
}

func TestFakeStructFillsTaggedFields(t *testing.T) {
	a, b := structData(t), structData(t)
	people := map[string]string{"Ada": "Lovelace", "Bo": "Ek"}
	zips := map[string]string{"Stockholm": "111 22", "Tranås": "573 31"}
	actives, nils := 0, 0
	for i := 0; i < 100; i++ {
		oldNick, oldTwinNick := "old", "old"
		u, twin := structUser{Nick: &oldNick, Note: "keep"}, structUser{Nick: &oldTwinNick, Note: "keep"}
		if err := a.FakeStruct(&u); err != nil {
			t.Fatal(err)
		}
		if err := b.FakeStruct(&twin); err != nil {
			t.Fatal(err)
		}
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
		if u.Active {
			actives++
		}
		if u.Nick == nil {
			nils++
		}
	}
	if actives == 0 || actives == 100 || nils == 0 || nils == 100 {
		t.Errorf("100 draws gave %d active and %d nil nicks, want both outcomes of each", actives, nils)
	}
}

func TestFakeStructFillsEmbeddedFieldsIntoItsRecord(t *testing.T) {
	f := structData(t)
	for i := 0; i < 100; i++ {
		var e structEmployee
		if err := f.FakeStruct(&e); err != nil {
			t.Fatal(err)
		}
		if e.StructFamily == nil || e.Last != "Lovelace" && e.Last != "Ek" || e.Email != strings.ToLower(e.First)+"@example.com" {
			t.Fatalf("%+v, %+v: want a promoted field's name read by the struct's own, the embedded pointer allocated", e, e.StructFamily)
		}
	}
	var skipped struct {
		structGiven `fake:"-"`
		Email       string `fake:"{lowercase(/person.first)}@example.com"`
	}
	if err := f.FakeStruct(&skipped); err != nil || skipped.First != "" || skipped.Email == "" {
		t.Errorf("FakeStruct = %v, %+v; want the embedded struct under fake:\"-\" left unfilled beside the filled field", err, skipped)
	}
}

func TestFakeStructDrawsANestedStructApart(t *testing.T) {
	f := structData(t)
	for i := 0; i < 100; i++ {
		var trip struct{ From, To structPlace }
		if err := f.FakeStruct(&trip); err != nil {
			t.Fatal(err)
		}
		if trip.From.City != trip.To.City {
			return
		}
	}
	t.Error("From and To drew one place in 100 trips; a nested struct is a record of its own, so each draws apart")
}

func TestFakeStructLeavesAPointerBackAlone(t *testing.T) {
	var l structLink
	if err := structData(t).FakeStruct(&l); err != nil || l.Name == "" || l.Next != nil {
		t.Errorf("FakeStruct = %v, %+v; want Name filled and Next, a pointer back to the struct being filled, left nil", err, l)
	}
}

func TestFakeStructReadsAPathInEachSpelling(t *testing.T) {
	v := struct {
		Path   string `fake:"person.first"`
		Ref    string `fake:"{/person.first}"`
		Quoted string `fake:"\"{/person.first}\""`
		Slash  string `fake:"/person.first"`
		Kept   string `fake:"-"`
	}{Kept: "kept"}
	if err := structData(t).FakeStruct(&v); err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string]string{"Path": v.Path, "Ref": v.Ref, "Quoted": v.Quoted, "Slash": v.Slash} {
		if got != "Ada" && got != "Bo" {
			t.Errorf("%s = %q, want a first name", name, got)
		}
	}
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

func TestFakeStructTypesABareNumberOrBooleanByItsGoType(t *testing.T) {
	f, seen := structData(t), map[bool]bool{}
	for i := 0; i < 30; i++ {
		v := struct {
			N int    `fake:"[5, 6]"`
			B bool   `fake:"[true, false]"`
			S string `fake:"[7.50]"`
		}{}
		if err := f.FakeStruct(&v); err != nil || (v.N != 5 && v.N != 6) || v.S != "7.50" {
			t.Fatalf("FakeStruct = %+v, %v, want each value typed by its field", v, err)
		}
		seen[v.B] = true
	}
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
		v    any
		want string
	}{
		{structUser{}, "fills a struct through a non-nil pointer, got fejkdata.structUser"},
		{(*structUser)(nil), "through a non-nil pointer"},
		{new(int), "through a non-nil pointer"},
		{nil, "through a non-nil pointer"},
		{&struct {
			a string `fake:"person.first"`
		}{}, ".a: unexported"},
		{&struct {
			A []string `fake:"person.first"`
		}{}, ".A: a fake tag fills a string, bool, integer or float field, or a pointer to one, not []string"},
		{&struct {
			A **int `fake:"{int(1,9)}"`
		}{}, "not **int"},
		{&struct {
			A structPlace `fake:"place"`
		}{}, ".A: a struct field fills from the tags on its own fields"},
		{&struct {
			A int `fake:"{\"format\":\"{int(1,9)}\",\"datatype\":\"number\"}"`
		}{}, `its Go type int sets the datatype integer; drop "datatype"`},
		{&struct {
			A string `fake:"{\"format\":\"1\",\"datatype\":\"integer\"}"`
		}{}, `its Go type string sets the datatype string; drop "datatype"`},
		{&struct {
			A int `fake:"[null,\"{int(1,9)}\"]"`
		}{}, "can draw null, which int cannot hold; make it *int"},
		{&struct {
			A int64 `fake:"src.score"`
		}{}, "can draw null, which int64 cannot hold; make it *int64"},
		{&struct {
			A string `fake:"src.code"`
		}{}, "can draw null, which string cannot hold; make it *string"},
		{&struct {
			A *bool `fake:"src.score"`
		}{}, ".A (*bool): {int(1,9)} prints an integer, not a boolean"},
		{&struct {
			A int64 `fake:"mid.score"`
		}{}, "can draw null, which int64 cannot hold; make it *int64"},
		{&struct {
			A int `fake:"{digits(3)}"`
		}{}, ".A (int): {digits(3)} prints text, not an integer"},
		{&struct {
			A int `fake:"person.first"`
		}{}, `"Ada" is not an integer`},
		{&struct {
			A int `fake:"{\"format\":\"{calc(d + 1)}\",\"d\":\"{digits(0)}\"}"`
		}{}, "{digits(0)} prints nothing, which is no number"},
		{&struct {
			A int8 `fake:"{int(0,300)}"`
		}{}, `"{int(0,300)}" is not proven within int8; narrow it to that range, or make the field int64`},
		{&struct {
			A uint `fake:"{int(-1,5)}"`
		}{}, `"{int(-1,5)}" is not proven within uint; narrow it to that range, or make the field int64`},
		{&struct {
			A float32 `fake:"[\"1\",\"1e39\"]"`
		}{}, `"1e39" is not proven within float32; narrow it to that range, or make the field float64`},
		{&struct {
			A int32 `fake:"{seq()}"`
		}{}, `"{seq()}" is not proven within int32; narrow it to that range, or make the field int64`},
		{&struct {
			A bool `fake:"{int(0,1)}"`
		}{}, "prints an integer, not a boolean"},
		{&struct {
			A string `fake:"nope.x"`
		}{}, `no entry "nope"`},
		{&struct {
			A string `fake:"a|b"`
		}{}, `contains "|"`},
		{&struct {
			A string `fake:""`
		}{}, "is empty"},
		{&struct {
			A string `fake:"[abc]"`
		}{}, `"["`},
		{&struct {
			A string `fake:"{.person.first} x"`
		}{}, "write {/person.first}"},
		{&struct{ *structGiven }{}, "an unexported embedded pointer field cannot be set"},
		{&struct {
			structGiven
			First string `fake:"person.last"`
		}{}, "struct.structGiven.First: hidden by another field named First"},
		{&struct {
			A string `fake:"{x}"`
		}{}, `no field "x"`},
		{&struct {
			Trip struct {
				A int `fake:"{digits(3)}"`
			}
		}{}, ": struct.Trip.A (int): {digits(3)} prints text"},
	} {
		err := f.FakeStruct(c.v)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("FakeStruct(%T) = %v, want an error containing %q", c.v, err, c.want)
			continue
		}
		if again := f.FakeStruct(c.v); again == nil || again.Error() != err.Error() {
			t.Errorf("FakeStruct(%T) again = %v, want the first call's error, %v", c.v, again, err)
		}
	}
}

func TestFakeStructBoundsTheStructsATypeReaches(t *testing.T) {
	f := structData(t)
	place := reflect.TypeOf(structPlace{})
	tree := place
	for depth := 1; depth <= 8; depth++ {
		tree = reflect.StructOf([]reflect.StructField{{Name: "L", Type: reflect.PointerTo(tree)}, {Name: "R", Type: reflect.PointerTo(tree)}})
	}
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
