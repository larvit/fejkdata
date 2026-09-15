package fejkdata

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type structPlace struct {
	City string `fake:"place.city"`
	Zip  string `fake:"place.zip"`
}

type structUser struct {
	Active bool   `fake:"[\"true\",\"false\"]"`
	Age    uint8  `fake:"{int(18,99)}"`
	Email  string `fake:"{lowercase(/person.first)}@example.com"`
	First  string `fake:"person.first"`
	Home   structPlace
	ID     int64   `fake:"{seq()}"`
	Last   string  `fake:"person.last"`
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
	First string `fake:"person.first"`
}

type StructFamily struct {
	Last string `fake:"person.last"`
}

type structEmployee struct {
	structGiven
	*StructFamily
	Email string `fake:"{lowercase(/person.first)}@example.com"`
}

type structLink struct {
	Name string `fake:"person.first"`
	Next *structLink
}

func structData(t *testing.T) *Generator {
	t.Helper()
	return newGenerator(t, writeData(t, map[string]string{
		"person": `[{"format":"{first} {last}","first":"Ada","last":"Lovelace"},{"format":"{first} {last}","first":"Bo","last":"Ek"}]`,
		"place":  `[{"format":"{city}","city":"Stockholm","zip":"111 22"},{"format":"{city}","city":"Tranås","zip":"573 31"}]`,
		"src":    `{"format":"","code":[null,"200","404"],"score":[null,{"format":"{int(1,9)}","datatype":"integer"}]}`,
		"trip":   `{"format":"","leg":[{"format":"{to}","to":"Oslo"},{"format":"{to}","to":"Rome"}]}`,
	}), WithSeed(1))
}

func TestFakeStructFieldTaggedWithAColumnIsThatColumn(t *testing.T) {
	f := structData(t)
	nils := map[string]int{}
	for i := 0; i < 200; i++ {
		var v struct {
			Code  *string `fake:"src.code"`
			Label string  `fake:"n={/src.score}"`
			Score *int64  `fake:"src.score"`
		}
		if err := f.FakeStruct(&v); err != nil {
			t.Fatal(err)
		}
		switch {
		case v.Code == nil:
			nils["code"]++
		case *v.Code != "200" && *v.Code != "404":
			t.Fatalf("Code = %q, want nil or a code, never a null's \"\"", *v.Code)
		}
		switch {
		case v.Score == nil && v.Label == "n=":
			nils["score"]++
		case v.Score == nil || *v.Score < 1 || *v.Score > 9 || v.Label != "n="+strconv.FormatInt(*v.Score, 10):
			t.Fatalf("%+v, want Score nil beside Label n=, or one digit in both", v)
		}
	}
	for _, name := range []string{"code", "score"} {
		if n := nils[name]; n == 0 || n == 200 {
			t.Errorf("%s was nil %d times in 200 fills, want both outcomes", name, n)
		}
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
			t.Fatalf("person fields %q %q %q, want one person drawn across the struct", u.First, u.Last, u.Email)
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
	people := map[string]string{"Ada": "Lovelace", "Bo": "Ek"}
	for i := 0; i < 100; i++ {
		var e structEmployee
		if err := f.FakeStruct(&e); err != nil {
			t.Fatal(err)
		}
		if e.StructFamily == nil || people[e.First] != e.Last || e.Email != strings.ToLower(e.First)+"@example.com" {
			t.Fatalf("%+v, %+v: want the promoted fields one person with the struct's own, the embedded pointer allocated", e, e.StructFamily)
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
		{&struct{ A string }{}, "has no fake tags"},
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
			A int `fake:"{\"format\":\"{int(1,9)}\",\"datatype\":\"integer\"}"`
		}{}, `its Go type int sets the datatype; drop "datatype"`},
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
			A int `fake:"{digits(3)}"`
		}{}, ".A (int): {digits(3)} prints text, not an integer"},
		{&struct {
			A int `fake:"person.first"`
		}{}, `"Ada" is not an integer`},
		{&struct {
			A int8 `fake:"{int(0,300)}"`
		}{}, `"{int(0,300)}" is not proven within int8; make it int64`},
		{&struct {
			A uint `fake:"{int(-1,5)}"`
		}{}, `"{int(-1,5)}" is not proven within uint; make it int64`},
		{&struct {
			A float32 `fake:"[\"1\",\"1e39\"]"`
		}{}, `"1e39" is not proven within float32; make it float64`},
		{&struct {
			A int32 `fake:"{seq()}"`
		}{}, `"{seq()}" is not proven within int32; make it int64`},
		{&struct {
			A bool `fake:"{int(0,1)}"`
		}{}, "prints an integer, not a boolean"},
		{&struct {
			A string `fake:"nope.x"`
		}{}, `no entry "nope"`},
		{&struct {
			A string `fake:"{/person.first}"`
		}{}, "is the path person.first written as a template; write person.first"},
		{&struct {
			A string `fake:"\"{/person.first}\""`
		}{}, "is the path person.first written as a template; write person.first"},
		{&struct {
			A string `fake:"a|b"`
		}{}, `contains "|"`},
		{&struct {
			A string `fake:""`
		}{}, "is empty"},
		{&struct {
			A string `fake:"[abc]"`
		}{}, `holds a "["`},
		{&struct {
			A string `fake:"{.person.first} x"`
		}{}, "write {/person.first}"},
		{&struct {
			A string `fake:"/person.first"`
		}{}, "write person.first"},
		{&struct {
			A string `fake:"-"`
		}{}, `fake:"-" leaves a struct field unfilled`},
		{&struct{ *structGiven }{}, "an unexported embedded pointer field cannot be set"},
		{&struct {
			structGiven
			First string `fake:"person.last"`
		}{}, "struct.structGiven.First: hidden by another field named First"},
		{&struct {
			A string `fake:"{x}"`
		}{}, `no field "x"`},
		{&struct {
			A string `fake:"trip.leg"`
			B string `fake:"trip.leg.to"`
		}{}, "reads a path into"},
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
