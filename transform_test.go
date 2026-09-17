package fejkdata

import "testing"

func TestTransforms(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"{uppercase(x)}|{lowercase(x)}|{ascii(x)}","x":"Åsa Ödegård-Nuñez"}`: "ÅSA ÖDEGÅRD-NUÑEZ|åsa ödegård-nuñez|Asa Odegard-Nunez",
		`{"format":"{lowercase(ascii(x))}","x":"Åsa"}`:                                  "asa",
		`{"format":"{ascii(x)}","x":"Ærø ß 日本"}`:                                        "AEro ss ",
	} {
		if got := mustRender(t, f, src); got != want {
			t.Errorf("render(%s) = %q, want %q", src, got, want)
		}
	}
}

func TestTransformReadsTheHeldDraw(t *testing.T) {
	f := engine(2)
	src := `{"format":"{p.first}={lowercase(p.first)}","p":[{"format":"{first}","first":"Anna"},{"format":"{first}","first":"Bo"}]}`
	for i := 0; i < 50; i++ {
		if got := mustRender(t, f, src); got != "Anna=anna" && got != "Bo=bo" {
			t.Fatalf("render = %q, want the transform over the same draw", got)
		}
	}
}

func TestTransformOverAReference(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `[{"format":"{first} {last}","first":"Åsa","last":"Öberg"},{"format":"{first} {last}","first":"Bo","last":"Ek","born":"1990"}]`,
		"email":  `"{/person.first} {/person.last} <{lowercase(ascii(/person.first))}.{lowercase(ascii(/person.last))}@example.com>"`,
	})
	f := newGenerator(t, dir, WithSeed(5))
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "email"); got != "Åsa Öberg <asa.oberg@example.com>" && got != "Bo Ek <bo.ek@example.com>" {
			t.Fatalf("email = %q, want a record from one draw", got)
		}
	}
}

func TestTransformArgs(t *testing.T) {
	for _, bad := range []string{
		`{"format":"{lowercase()}","x":"v"}`,
		`{"format":"{lowercase(nope)}","x":"v"}`,
		`{"format":"{ascii(x,y)}","x":"v","y":"w"}`,
		`{"format":"{lowercase(hex(2))}","x":"v"}`,
		`{"format":"{lowercase(x)} {x.a}","x":{"format":"{a}","a":"1"}}`,
	} {
		if _, err := compile(parse(t, bad)); err == nil {
			t.Errorf("compile(%s) = nil error, want it rejected", bad)
		}
	}
}

func TestAsciiKeepsDEL(t *testing.T) {
	if got := mustRender(t, engine(1), `{"format":"{ascii(x)}","x":"a\u007fb"}`); got != "a\u007fb" {
		t.Errorf("ascii over DEL = %q, want it kept: DEL is ASCII", got)
	}
}
