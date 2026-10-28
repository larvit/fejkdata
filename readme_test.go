package fejkdata

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	svse "github.com/larvit/fejkdata/data/sv_SE"
)

var (
	jsonBlock  = regexp.MustCompile("(?s)```json\n(.*?)```")
	plainBlock = regexp.MustCompile("(?s)```\n(.*?)```")
	tsvBlock   = regexp.MustCompile("(?s)```tsv\n(.*?)```")
	goBlock    = regexp.MustCompile("(?s)```go\n(.*?)```")
	textBlock  = regexp.MustCompile("(?s)```text\n(.*?)```")
	rowsFile   = regexp.MustCompile(`"rows":\s*"([^"]+)"`)
	parentName = regexp.MustCompile(`"parent":\s*"([^"]+)"`)
	headingEnd = regexp.MustCompile(`\n##+ `)
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
		f, err := New(withShipped(), WithDataPath(writeFiles(t, exampleFiles(t, src, at[0], body))), WithSeed(1))
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
	f, err := New(withShipped(), WithDataPath(writeData(t, map[string]string{"users": block[1]})), WithSeed(1))
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
	f, err := New(withShipped(), WithDataPath(writeData(t, map[string]string{"sql": block[1]})), WithSeed(1))
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
	f, err := New(withShipped(), WithDataPath(writeData(t, map[string]string{"order": block[1]})), WithSeed(1))
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
	f, err := New(withShipped(), WithDataPath(writeFiles(t, exampleFiles(t, src, i+at[0], body))), WithSeed(1))
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
	for _, name := range packageDir(t).GoFiles {
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

func shCommands(src string) []string {
	var commands []string
	for _, block := range regexp.MustCompile("(?s)```sh\n(.*?)```").FindAllStringSubmatch(src, -1) {
		for _, line := range strings.Split(strings.TrimSpace(block[1]), "\n") {
			command, _, _ := strings.Cut(line, "#")
			commands = append(commands, strings.TrimSpace(command))
		}
	}
	return commands
}

func TestDockerhubPageSaysOnlyWhatReadmeSays(t *testing.T) {
	doc := readme(t)
	description, err := os.ReadFile("dockerhub/description.txt")
	if err != nil {
		t.Fatal(err)
	}
	tagline := strings.TrimSpace(string(description))
	if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(tagline) + `[:.]`).MatchString(doc) {
		t.Errorf("dockerhub/description.txt says %q, which no README.md sentence opens with", tagline)
	}
	overview, err := os.ReadFile("dockerhub/overview.md")
	if err != nil {
		t.Fatal(err)
	}
	commands := shCommands(string(overview))
	if len(commands) == 0 {
		t.Fatal("dockerhub/overview.md has no sh block")
	}
	shown := shCommands(doc)
	docker := regexp.MustCompile(`docker run --rm (-i )?larvit/fejkdata`)
	for _, command := range commands {
		before, after, _ := strings.Cut(command, "docker run")
		reads := strings.Contains(before, "|") || strings.Contains(after, "<")
		if reads != strings.Contains(command, "docker run --rm -i ") {
			t.Errorf("dockerhub/overview.md runs %q, which needs -i exactly when it reads stdin", command)
		}
		if !slices.Contains(shown, docker.ReplaceAllString(command, "fejkdata")) {
			t.Errorf("dockerhub/overview.md runs %q, which no README.md sh line shows", command)
		}
	}
}

// readmeSection is the README from heading to the next heading.
func readmeSection(t *testing.T, src, heading string) string {
	t.Helper()
	i := strings.Index(src, heading+"\n")
	if i < 0 {
		t.Fatalf("README lost the %q section", heading)
	}
	section := src[i+len(heading):]
	if end := headingEnd.FindStringIndex(section); end != nil {
		section = section[:end[0]]
	}
	return section
}

func tsvRows(t *testing.T, path string) []map[string]string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(src)), "\n")
	head := strings.Split(lines[0], "\t")
	rows := make([]map[string]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		cells := strings.Split(line, "\t")
		row := make(map[string]string, len(head))
		for i, name := range head {
			row[name] = cells[i]
		}
		rows = append(rows, row)
	}
	return rows
}

// quickStartPerson is one person as a quick start's columns or fields name it.
type quickStartPerson struct {
	sex, first, last, personnummer, email, street, postalCode, locality string
}

// quickStartSex holds each sex's code in first-name.tsv and the birth number Skatteverket's test series gives it.
var quickStartSex = map[string]struct{ code, birthNumber string }{
	"kvinna": {"f", "238"},
	"man":    {"m", "239"},
}

var (
	asciiFold         = strings.NewReplacer("å", "a", "ä", "a", "ã", "a", "ö", "o", "ø", "o", "ó", "o", "é", "e", "ü", "u")
	slugRun           = regexp.MustCompile(`[^a-z0-9]+`)
	personnummerShape = regexp.MustCompile(`^(\d{6})-(\d{3})(\d)$`)
)

// emailSlug is a name as the quick starts' email slugs it: folded, lower case, every run of other characters one `-`, none at either end.
func emailSlug(name string) string {
	return strings.Trim(slugRun.ReplaceAllString(asciiFold.Replace(strings.ToLower(name)), "-"), "-")
}

// quickStartShipped holds what a quick start's person must agree with: the first names shipped per sex code, and the geo tables geoPlaces reads.
type quickStartShipped struct {
	firstNames map[string]map[string]bool
	places     geoPlaceSet
}

func loadQuickStartShipped(t *testing.T) quickStartShipped {
	t.Helper()
	s := quickStartShipped{firstNames: map[string]map[string]bool{}, places: geoPlaces(t, "SE")}
	for _, row := range tsvRows(t, "data/sv_SE/first-name.tsv") {
		if s.firstNames[row["sex"]] == nil {
			s.firstNames[row["sex"]] = map[string]bool{}
		}
		s.firstNames[row["sex"]][row["name"]] = true
	}
	return s
}

// checkPerson says how p's columns fail to agree, or returns nil.
func (s quickStartShipped) checkPerson(p quickStartPerson) error {
	sex, ok := quickStartSex[p.sex]
	if !ok {
		return fmt.Errorf("sex %q is not one the shipped names carry", p.sex)
	}
	if !s.firstNames[sex.code][p.first] {
		return fmt.Errorf("first name %q is not shipped for sex %q", p.first, p.sex)
	}
	want := emailSlug(p.first) + "." + emailSlug(p.last) + "@example.com"
	if p.email != want {
		return fmt.Errorf("email %q, want %q", p.email, want)
	}
	pnr := personnummerShape.FindStringSubmatch(p.personnummer)
	if pnr == nil {
		return fmt.Errorf("personnummer %q is not YYMMDD-NNNC", p.personnummer)
	}
	if pnr[2] != sex.birthNumber {
		return fmt.Errorf("personnummer %q has birth number %s, sex %q takes %s", p.personnummer, pnr[2], p.sex, sex.birthNumber)
	}
	century := "20"
	if pnr[1][:2] >= "40" {
		century = "19"
	}
	day, err := time.Parse("20060102", century+pnr[1])
	if err != nil || day.Before(time.Date(1940, 1, 1, 0, 0, 0, 0, time.UTC)) || day.After(time.Date(2006, 12, 31, 0, 0, 0, 0, time.UTC)) {
		return fmt.Errorf("personnummer %q has a birth date outside 1940-01-01 to 2006-12-31", p.personnummer)
	}
	if !luhnValid(pnr[1] + pnr[2] + pnr[3]) {
		return fmt.Errorf("personnummer %q fails the Luhn check", p.personnummer)
	}
	number := strings.LastIndex(p.street, " ")
	if number < 0 {
		return fmt.Errorf("street %q has no number", p.street)
	}
	if !s.places.has(p.street[:number] + "|" + p.postalCode + "|" + p.locality + "|") {
		return fmt.Errorf("street %q, postal code %q and locality %q do not belong together", p.street, p.postalCode, p.locality)
	}
	return nil
}

// readmeStruct returns the struct the README's Go block declares as Person, with the tags it writes.
func readmeStruct(t *testing.T, block string) reflect.Type {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "README.md", block, 0)
	if err != nil {
		t.Fatalf("README's Go quick start does not parse: %v", err)
	}
	var fields []reflect.StructField
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "Person" {
			return true
		}
		st, ok := spec.Type.(*ast.StructType)
		if !ok {
			t.Fatal("README's Person is not a struct")
		}
		for _, field := range st.Fields.List {
			if id, ok := field.Type.(*ast.Ident); !ok || id.Name != "string" {
				t.Fatal("README's Person has a field that is not a string")
			}
			tag := ""
			if field.Tag != nil {
				tag, _ = strconv.Unquote(field.Tag.Value)
			}
			for _, name := range field.Names {
				fields = append(fields, reflect.StructField{Name: name.Name, Type: reflect.TypeOf(""), Tag: reflect.StructTag(tag)})
			}
		}
		return false
	})
	if fields == nil {
		t.Fatal("README's Go quick start declares no Person struct")
	}
	return reflect.StructOf(fields)
}

func TestReadmeBulkQuickStart(t *testing.T) {
	t.Parallel()
	section := readmeSection(t, readme(t), "### Bulk fixtures")
	block := jsonBlock.FindStringSubmatch(section)
	if block == nil {
		t.Fatal("README's bulk quick start has no json record")
	}
	dir := writeData(t, map[string]string{"person": block[1]})
	f, err := New(withShipped(), WithDataPath(dir), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	shipped := loadQuickStartShipped(t)
	for i := 0; i < 300; i++ {
		r, err := f.FakeRecord("person")
		if err != nil {
			t.Fatal(err)
		}
		col := map[string]string{}
		for _, c := range r.Columns() {
			col[c.Name] = c.Value
		}
		p := quickStartPerson{col["sex"], col["first"], col["last"], col["personnummer"], col["email"], col["street"], col["postal-code"], col["locality"]}
		if err := shipped.checkPerson(p); err != nil {
			t.Fatalf("%v: %v", col, err)
		}
	}
}

func TestReadmeBulkQuickStartOutput(t *testing.T) {
	section := readmeSection(t, readme(t), "### Bulk fixtures")
	block := jsonBlock.FindStringSubmatch(section)
	out := textBlock.FindStringSubmatch(section)
	if block == nil || out == nil {
		t.Fatal("README's bulk quick start lost its record or its printed output")
	}
	if !strings.Contains(section, "fejkdata --seed 1 --format csv < person.json") {
		t.Fatal("README's bulk quick start no longer runs the command its output shows")
	}
	f, err := New(withShipped(), WithDataPath(writeData(t, map[string]string{"person": block[1]})), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	r, err := f.FakeRecord("person")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.CSVHeader()+"\n"+r.CSVLine(), strings.TrimSpace(out[1]); got != want {
		t.Errorf("README prints %q for --seed 1 --format csv, it renders %q", want, got)
	}
}

func TestReadmeGoQuickStart(t *testing.T) {
	t.Parallel()
	section := readmeSection(t, readme(t), "### Go tests")
	block := goBlock.FindStringSubmatch(section)
	printed := regexp.MustCompile(`(?m)^\s*// With seed 1, p is (\{.*\})$`).FindStringSubmatch(section)
	if block == nil || printed == nil {
		t.Fatal("README's Go quick start lost its go block or its printed person")
	}
	if !strings.Contains(block[1], "fejkdata.New(fejkdata.WithDataFS(svse.Modules()...), fejkdata.WithSeed(1))") {
		t.Error("README's Go quick start no longer has the New call this test runs")
	}
	if !strings.Contains(block[1], "f.FakeStruct(&p)") {
		t.Error("README's Go quick start no longer has the FakeStruct call this test runs")
	}
	if !strings.Contains(block[1], `"github.com/larvit/fejkdata/data/sv_SE"`) {
		t.Error("README's Go quick start no longer imports the data this test loads")
	}
	typ := readmeStruct(t, block[1])
	f, err := New(WithDataFS(svse.Modules()...), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	first := reflect.New(typ)
	if err := f.FakeStruct(first.Interface()); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%+v", first.Elem().Interface()); got != printed[1] {
		t.Errorf("README prints %s for seed 1, it renders %s", printed[1], got)
	}
	shipped := loadQuickStartShipped(t)
	for i := 0; i < 300; i++ {
		ptr := reflect.New(typ)
		if err := f.FakeStruct(ptr.Interface()); err != nil {
			t.Fatal(err)
		}
		field := func(name string) string {
			v := ptr.Elem().FieldByName(name)
			if !v.IsValid() {
				t.Fatalf("README's Go quick start has no %s field", name)
			}
			return v.String()
		}
		p := quickStartPerson{field("Sex"), field("First"), field("Last"), field("Personnummer"), field("Email"), field("Street"), field("PostalCode"), field("Locality")}
		if err := shipped.checkPerson(p); err != nil {
			t.Fatalf("%+v: %v", p, err)
		}
	}
}
