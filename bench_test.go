package fejkdata

import (
	"os"
	"path/filepath"
	"testing"
)

func benchGenerator(b *testing.B, data Option) *Generator {
	b.Helper()
	f, err := New(data, WithSeed(1))
	if err != nil {
		b.Fatal(err)
	}
	return f
}

func benchPath(b *testing.B, data Option, path string) {
	f := benchGenerator(b, data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := f.Fake(path); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPerson(b *testing.B)       { benchPath(b, withShipped(), "sv_SE.person") }
func BenchmarkAddress(b *testing.B)      { benchPath(b, withShipped(), "sv_SE.address") }
func BenchmarkWord(b *testing.B)         { benchPath(b, withShipped(), "sv_SE.word") }
func BenchmarkCreditcard(b *testing.B)   { benchPath(b, withShipped(), "misc.creditcard") }
func BenchmarkPersonnummer(b *testing.B) { benchPath(b, withShipped(), "sv_SE.personnummer") }
func BenchmarkUUIDv7(b *testing.B)       { benchPath(b, withShipped(), "misc.uuid") }

func tmpData(b *testing.B, name, body string) string {
	b.Helper()
	dir := b.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(body), 0o644); err != nil {
		b.Fatal(err)
	}
	return dir
}

func BenchmarkCalc(b *testing.B) {
	dir := tmpData(b, "inv", `{"format":"{net} x {qty} = {calc(net * qty, 2)}","net":"19.99","qty":"3"}`)
	benchPath(b, WithDataPath(dir), "inv")
}

func BenchmarkLongLiteral(b *testing.B) {
	dir := tmpData(b, "sql", `{"format":"INSERT INTO customers (id, name, city) VALUES (123, '{word}', '{word}');","word":["alpha","beta","gamma","delta"]}`)
	benchPath(b, WithDataPath(dir), "sql")
}

func BenchmarkRepeat(b *testing.B) {
	dir := tmpData(b, "many", `{"format":"{word}","repeat":20,"separator":", ","word":["alpha","beta","gamma","delta"]}`)
	benchPath(b, WithDataPath(dir), "many")
}

// BenchmarkNamed measures a format that binds a field to a name and reads two paths through
// it — what a correlated pair costs against BenchmarkUnnamed, the same output drawn from two
// independent fields.
func BenchmarkNamed(b *testing.B) {
	dir := tmpData(b, "addr", `{"format":"{place as p}{p.postal-code} {p.locality}","place":[{"format":"{locality}","locality":"Stockholm","postal-code":"1{digits(2)} {digits(2)}"},{"format":"{locality}","locality":"Tranås","postal-code":"573 {digits(2)}"}]}`)
	benchPath(b, WithDataPath(dir), "addr")
}

func BenchmarkUnnamed(b *testing.B) {
	dir := tmpData(b, "addr", `{"format":"{postal-code} {locality}","postal-code":["1{digits(2)} {digits(2)}","573 {digits(2)}"],"locality":["Stockholm","Tranås"]}`)
	benchPath(b, WithDataPath(dir), "addr")
}

// BenchmarkNamedDeep reads two paths through a name and an intermediate level.
func BenchmarkNamedDeep(b *testing.B) {
	dir := tmpData(b, "addr", `{"format":"{p as q}{q.addr.city} {q.addr.zip}","p":{"format":"{addr}","addr":{"format":"{city}","city":["Stockholm","Tranås"],"zip":"1{digits(2)} {digits(2)}"}}}`)
	benchPath(b, WithDataPath(dir), "addr")
}

// BenchmarkNamedWide reads ten paths through one name: the point where the maps a pick keeps
// outgrow a single bucket.
func BenchmarkNamedWide(b *testing.B) {
	dir := tmpData(b, "row", `{"format":"{r as s}{s.a}{s.b}{s.c}{s.d}{s.e}{s.f}{s.g}{s.h}{s.i}{s.j}","r":[
		{"format":"x","a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"0"},
		{"format":"y","a":"A","b":"B","c":"C","d":"D","e":"E","f":"F","g":"G","h":"H","i":"I","j":"J"}]}`)
	benchPath(b, WithDataPath(dir), "row")
}

func BenchmarkFirstFake(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f, err := New(withShipped(), WithSeed(1))
		if err != nil {
			b.Fatal(err)
		}
		if _, err := f.Fake("sv_SE.address"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := New(withShipped(), WithSeed(1)); err != nil {
			b.Fatal(err)
		}
	}
}
