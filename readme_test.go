package fejkdata

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var (
	jsonBlock  = regexp.MustCompile("(?s)```json\n(.*?)```")
	plainBlock = regexp.MustCompile("(?s)```\n(.*?)```")
	tsvBlock   = regexp.MustCompile("(?s)```tsv\n(.*?)```")
	rowsFile   = regexp.MustCompile(`"rows":\s*"([^"]+)"`)
	parentName = regexp.MustCompile(`"parent":\s*"([^"]+)"`)
)

// exampleFiles is a README json block as a data directory's files: the category, the
// rows TSV it names, taken from the nearest tsv block above it, and the parent table it
// names, taken from the nearest json block above it whose rows file is the parent's.
func exampleFiles(t *testing.T, src string, at int, body string) map[string]string {
	t.Helper()
	files := map[string]string{"example.json": body}
	if m := rowsFile.FindStringSubmatch(body); m != nil {
		tsv := tsvBlock.FindAllStringSubmatch(src[:at], -1)
		if tsv == nil {
			t.Fatalf("README example names %s with no tsv block above it", m[1])
		}
		files[m[1]] = tsv[len(tsv)-1][1]
	}
	if m := parentName.FindStringSubmatch(body); m != nil {
		blocks := jsonBlock.FindAllStringSubmatchIndex(src[:at], -1)
		for i := len(blocks) - 1; i >= 0; i-- {
			parent := src[blocks[i][2]:blocks[i][3]]
			if strings.Contains(parent, `"rows": "`+m[1]+`.tsv"`) {
				for name, content := range exampleFiles(t, src, blocks[i][0], parent) {
					if name == "example.json" {
						name = m[1] + ".json"
					}
					files[name] = content
				}
				break
			}
		}
	}
	return files
}

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
	src := readme(t)
	for _, at := range jsonBlock.FindAllStringSubmatchIndex(src, -1) {
		body := src[at[2]:at[3]]
		if strings.Contains(body, "…") {
			continue
		}
		f, err := New(WithDataPath(writeFiles(t, exampleFiles(t, src, at[0], body))), WithSeed(1))
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

func TestReadmeTableExample(t *testing.T) {
	src := readme(t)
	i := strings.Index(src, "### Table")
	if i < 0 {
		t.Fatal("README lost the Table section")
	}
	at := jsonBlock.FindStringSubmatchIndex(src[i:])
	if at == nil {
		t.Fatal("README Table section has no json block")
	}
	body := src[i:][at[2]:at[3]]
	f, err := New(WithDataPath(writeFiles(t, exampleFiles(t, src, i+at[0], body))), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{"example[SE]": "Sweden (SE)", "example[Norway].alpha2": "NO"} {
		if got := fake(t, f, path); got != want {
			t.Errorf("README table example: Fake(%q) = %q, want %q", path, got, want)
		}
	}
	r, err := f.FakeRecord("example")
	if err != nil || r.CSVHeader() != "alpha2,name,population" {
		t.Fatalf("README table example as a record: %q, %v", r.CSVHeader(), err)
	}
}

// TestLayoutNamesEverySourceFile holds an added, renamed or split file to its Layout line.
func TestLayoutNamesEverySourceFile(t *testing.T) {
	src := readme(t)
	i := strings.Index(src, "## Layout")
	if i < 0 {
		t.Fatal("README lost the Layout section")
	}
	block := plainBlock.FindStringSubmatch(src[i:])
	if block == nil {
		t.Fatal("README Layout section has no block")
	}
	listed := map[string]bool{}
	for _, line := range strings.Split(block[1], "\n") {
		name, _, _ := strings.Cut(strings.TrimSpace(line), " ")
		if strings.HasSuffix(name, ".go") {
			listed[name] = true
		}
	}
	for _, name := range goFiles(t) {
		if !listed[name] {
			t.Errorf("the README Layout names no %s, so nothing says what it holds", name)
		}
		delete(listed, name)
	}
	stale := make([]string, 0, len(listed))
	for name := range listed {
		stale = append(stale, name)
	}
	slices.Sort(stale)
	for _, name := range stale {
		t.Errorf("the README Layout names %s, which the package does not hold", name)
	}
}
