package fejkdata

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

var jsonBlock = regexp.MustCompile("(?s)```json\n(.*?)```")

func readme(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

func TestReadmeExamplesLoadAndRender(t *testing.T) {
	n := 0
	for _, m := range jsonBlock.FindAllStringSubmatch(readme(t), -1) {
		body := m[1]
		if strings.Contains(body, "…") {
			continue
		}
		f, err := New(WithDataPath(writeData(t, map[string]string{"example": body})), WithSeed(1))
		if err != nil {
			t.Errorf("README example does not load: %v\n%s", err, body)
			continue
		}
		for i := 0; i < 20; i++ {
			if _, err := f.Fake("example"); err != nil {
				t.Errorf("README example does not render: %v\n%s", err, body)
				break
			}
		}
		n++
	}
	if n < 8 {
		t.Fatalf("found only %d README examples", n)
	}
}

func TestReadmeRecordExample(t *testing.T) {
	src := readme(t)
	section := src[strings.Index(src, "### Records"):]
	section = section[:strings.Index(section, "## Library")]
	block := jsonBlock.FindStringSubmatch(section)
	if block == nil {
		t.Fatal("README lost the Records example")
	}
	f, err := New(WithDataPath(writeData(t, map[string]string{"users": block[1]})), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	r, err := f.FakeRecord("users")
	if err != nil {
		t.Fatal(err)
	}
	got := r.Columns()
	if len(got) != 2 || got[0].Name != "first" || got[0].Value != "Bo" || got[1].Name != "last" || got[1].Value != "Lovelace" {
		t.Fatalf("record columns = %v, want first=Bo, last=Lovelace with seed 1", got)
	}
	if r.CSVHeader() != "first,last" || r.CSVLine() != "Bo,Lovelace" {
		t.Fatalf("csv = %q, %q, want first,last / Bo,Lovelace", r.CSVHeader(), r.CSVLine())
	}
	if r.SQLInsert("users") != `INSERT INTO "users" ("first", "last") VALUES ('Bo', 'Lovelace');` {
		t.Fatalf("sql = %q, want the seeded INSERT", r.SQLInsert("users"))
	}
}

func TestReadmeSQLExampleOutput(t *testing.T) {
	src := readme(t)
	src = src[strings.Index(src, "### Your own data"):]
	block := jsonBlock.FindStringSubmatch(src)
	want := regexp.MustCompile(`(?m)^# (INSERT .*)$`).FindStringSubmatch(src)
	if block == nil || want == nil {
		t.Fatal("README lost the SQL example or its printed output")
	}
	f, err := New(WithDataPath(writeData(t, map[string]string{"sql": block[1]})), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	if got := fake(t, f, "sql"); got != want[1] {
		t.Errorf("README SQL example with seed 1 = %q, README prints %q", got, want[1])
	}
}

func TestReadmeDatatypeExample(t *testing.T) {
	src := readme(t)
	i := strings.Index(src, "### Datatype")
	if i < 0 {
		t.Fatal("README lost the Datatype section")
	}
	block := jsonBlock.FindStringSubmatch(src[i:])
	f, err := New(WithDataPath(writeData(t, map[string]string{"order": block[1]})), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	r, err := f.FakeRecord("order")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(r.JSON()), &m); err != nil {
		t.Fatalf("JSON() = %s: %v", r.JSON(), err)
	}
	shown := map[DataType]bool{}
	for _, c := range r.Columns() {
		shown[c.DataType] = true
		var ok bool
		switch c.DataType {
		case DataTypeBoolean:
			_, ok = m[c.Name].(bool)
		case DataTypeInteger, DataTypeNumber:
			_, ok = m[c.Name].(float64)
		default:
			_, ok = m[c.Name].(string)
		}
		if !ok {
			t.Errorf("column %q, datatype %s, written as %s", c.Name, c.DataType, r.JSON())
		}
	}
	if !shown[DataTypeInteger] || !shown[DataTypeNumber] || !shown[DataTypeBoolean] {
		t.Errorf("README Datatype example shows %v, want an integer, a number and a boolean column", shown)
	}
}
