package fejkdata

import (
	"os"
	"path/filepath"
	"testing"
)

func columnNames(r *Record) []string {
	var names []string
	for _, c := range r.Columns() {
		names = append(names, c.Name)
	}
	return names
}

func TestRecordColumnsFollowTheOrderTheTemplateListsItsFields(t *testing.T) {
	dir := writeData(t, map[string]string{
		"place": `{"zip": "1", "format": "", "city": "2", "hidden": {"secret": "4"}, "age": "3"}`,
	})
	r, err := newGenerator(t, dir, WithSeed(1)).FakeRecord("place")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.JSON(), `{"zip":"1","city":"2","age":"3"}`; got != want {
		t.Errorf("JSON() = %s, want %s", got, want)
	}
	if got, want := r.CSVHeader(), "zip,city,age"; got != want {
		t.Errorf("CSVHeader() = %s, want %s", got, want)
	}
	if got, want := r.CSVLine(), "1,2,3"; got != want {
		t.Errorf("CSVLine() = %s, want %s", got, want)
	}
	if got, want := r.SQLInsert("place"), `INSERT INTO "place" ("zip", "city", "age") VALUES ('1', '2', '3');`; got != want {
		t.Errorf("SQLInsert() = %s, want %s", got, want)
	}
}

func TestInlineRecordColumnsFollowTheOrderTheTemplateListsItsFields(t *testing.T) {
	r, err := recordCat(t).FakeRecordTemplate(`{"format":"","b":"1","a":{"format":"{y}{x}","y":"2","x":"3"},"c":"4"}`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.JSON(), `{"b":"1","a":"23","c":"4"}`; got != want {
		t.Errorf("JSON() = %s, want %s", got, want)
	}
}

func TestRecordColumnOfAKeyWrittenTwiceTakesItsFirstPlaceAndItsLastValue(t *testing.T) {
	r, err := recordCat(t).FakeRecordTemplate(`{"format":"","b":"1","a":"2","b":"3"}`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.JSON(), `{"b":"3","a":"2"}`; got != want {
		t.Errorf("JSON() = %s, want %s", got, want)
	}
}

func TestRecordDrawsTheSameValuesInWhateverOrderItsFieldsAreWritten(t *testing.T) {
	const digits = `["0","1","2","3","4","5","6","7","8","9"]`
	inNameOrder := `{"format":"","a":` + digits + `,"b":` + digits + `,"c":` + digits + `}`
	reversed := `{"format":"","c":` + digits + `,"b":` + digits + `,"a":` + digits + `}`
	for seed := uint64(1); seed <= 20; seed++ {
		want, err := newGenerator(t, writeData(t, nil), WithSeed(seed)).FakeRecordTemplate(inNameOrder)
		if err != nil {
			t.Fatal(err)
		}
		got, err := newGenerator(t, writeData(t, nil), WithSeed(seed)).FakeRecordTemplate(reversed)
		if err != nil {
			t.Fatal(err)
		}
		if g, w := valuesByColumn(got), valuesByColumn(want); g["a"] != w["a"] || g["b"] != w["b"] || g["c"] != w["c"] {
			t.Fatalf("seed %d: reversed fields drew %v, want the values %v of the same fields in name order", seed, g, w)
		}
	}
}

func TestTableRecordColumnsFollowItsHeader(t *testing.T) {
	dir := writeData(t, map[string]string{"town": `{"format":"{name}","rows":"town.tsv","key":"code"}`})
	if err := os.WriteFile(filepath.Join(dir, "town.tsv"), []byte("name\tcode\tarea\nUmeå\t1\t9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := newGenerator(t, dir, WithSeed(1)).FakeRecord("town")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.CSVHeader(), "name,code,area"; got != want {
		t.Errorf("CSVHeader() = %s, want %s", got, want)
	}
	if got, want := r.CSVLine(), "Umeå,1,9"; got != want {
		t.Errorf("CSVLine() = %s, want %s", got, want)
	}
}
