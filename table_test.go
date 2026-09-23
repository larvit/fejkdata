package fejkdata

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// writeFiles writes a data directory from a map of relative file name, extension
// included, to content.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// geo is three linked tables: region <- municipality <- locality.
func geo() map[string]string {
	return map[string]string{
		"region.json":       `{"format":"{name}","rows":"region.tsv","key":"code","name":"name","weight":"population"}`,
		"region.tsv":        "code\tname\tpopulation\ttimezone\n01\tStockholms län\t2400000\tEurope/Stockholm\n12\tSkåne län\t1400000\tEurope/Stockholm\n14\tVästra Götalands län\t1750000\tEurope/Stockholm\n",
		"municipality.json": `{"format":"{name}","rows":"municipality.tsv","key":"code","name":"name","parent":"region","weight":"population"}`,
		"municipality.tsv":  "code\tname\tregion\tpopulation\n0180\tStockholm\t01\t980000\n0184\tSolna\t01\t85000\n1280\tMalmö\t12\t360000\n1281\tLund\t12\t130000\n1480\tGöteborg\t14\t590000\n",
		"locality.json":     `{"format":"{name}","rows":"locality.tsv","key":"code","name":"name","parent":"municipality"}`,
		"locality.tsv":      "code\tname\tmunicipality\nL1\tStockholm\t0180\nL2\tSolna\t0184\nL3\tMalmö\t1280\nL4\tLund\t1281\nL5\tGöteborg\t1480\nL6\tHisingen\t1480\nL7\tSandby\t1281\nL8\tSandby\t0184\n",
	}
}

var (
	municipalityOf = map[string]string{"L1": "0180", "L2": "0184", "L3": "1280", "L4": "1281", "L5": "1480", "L6": "1480", "L7": "1281", "L8": "0184"}
	regionOf       = map[string]string{"0180": "01", "0184": "01", "1280": "12", "1281": "12", "1480": "14"}
)

func siblings() map[string]string {
	return with(geo(), map[string]string{
		"postal-code.json": `{"format":"{code}","rows":"postal-code.tsv","key":"code","parent":"locality"}`,
		"postal-code.tsv":  "code\tlocality\n111 20\tL1\n111 21\tL1\n171 41\tL2\n211 20\tL3\n221 00\tL4\n223 50\tL4\n411 01\tL5\n417 05\tL6\n247 45\tL7\n170 71\tL8\n",
		"street.json":      `{"format":"{name}","rows":"street.tsv","parent":"locality","weight":"segments"}`,
		"street.tsv":       "name\tlocality\tsegments\nDrottninggatan\tL1\t30\nSveavägen\tL1\t12\nRåsundavägen\tL2\t8\nStorgatan\tL3\t20\nStora Södergatan\tL4\t9\nKlostergatan\tL4\t4\nAvenyn\tL5\t15\nHisingsgatan\tL6\t3\nSandbyvägen\tL7\t2\nSandbyvägen\tL8\t2\n",
	})
}

var (
	localityOfCode   = map[string]string{"111 20": "L1", "111 21": "L1", "171 41": "L2", "211 20": "L3", "221 00": "L4", "223 50": "L4", "411 01": "L5", "417 05": "L6", "247 45": "L7", "170 71": "L8"}
	localityOfStreet = map[string][]string{"Drottninggatan": {"L1"}, "Sveavägen": {"L1"}, "Råsundavägen": {"L2"}, "Storgatan": {"L3"}, "Stora Södergatan": {"L4"}, "Klostergatan": {"L4"}, "Avenyn": {"L5"}, "Hisingsgatan": {"L6"}, "Sandbyvägen": {"L7", "L8"}}
)

func with(files map[string]string, more map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range files {
		out[k] = v
	}
	for k, v := range more {
		out[k] = v
	}
	return out
}

func TestTableRendersRowsAndColumns(t *testing.T) {
	f := newGenerator(t, writeFiles(t, geo()), WithSeed(1))
	names := map[string]bool{"Stockholms län": true, "Skåne län": true, "Västra Götalands län": true}
	for i := 0; i < 50; i++ {
		if v := fake(t, f, "region"); !names[v] {
			t.Fatalf("region = %q, want a row's format", v)
		}
		if v := fake(t, f, "region.timezone"); v != "Europe/Stockholm" {
			t.Fatalf("region.timezone = %q", v)
		}
		if v := fake(t, f, "region.code"); v != "01" && v != "12" && v != "14" {
			t.Fatalf("region.code = %q", v)
		}
	}
	want := []string{
		"locality", "locality.code", "locality.municipality", "locality.name",
		"municipality", "municipality.code", "municipality.locality", "municipality.locality.code", "municipality.locality.municipality", "municipality.locality.name", "municipality.name", "municipality.population", "municipality.region",
		"region", "region.code", "region.municipality", "region.municipality.code", "region.municipality.locality", "region.municipality.locality.code", "region.municipality.locality.municipality", "region.municipality.locality.name", "region.municipality.name", "region.municipality.population", "region.municipality.region", "region.name", "region.population", "region.timezone",
	}
	if got := f.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %v\nwant %v", got, want)
	}
	r, err := f.FakeRecord("region")
	if err != nil {
		t.Fatal(err)
	}
	if r.CSVHeader() != "code,name,population,timezone" {
		t.Fatalf("CSVHeader = %q", r.CSVHeader())
	}
	line := strings.Split(r.CSVLine(), ",")
	if len(line) != 4 || !names[line[1]] || line[3] != "Europe/Stockholm" {
		t.Fatalf("CSVLine = %q, want one row's cells", r.CSVLine())
	}
	for _, c := range r.Columns() {
		if c.DataType != DataTypeString || c.Null {
			t.Fatalf("column %q is %s null=%v, want a string", c.Name, c.DataType, c.Null)
		}
	}
}

func TestTableWeightSkewsTheDraw(t *testing.T) {
	f := newGenerator(t, writeFiles(t, geo()), WithSeed(3))
	count := map[string]int{}
	for i := 0; i < 3000; i++ {
		count[fake(t, f, "region.code")]++
	}
	if count["01"] < 1200 || count["12"] > 900 {
		t.Fatalf("region draws %v, want weighted by population (01 ≈ 43%%, 12 ≈ 25%%)", count)
	}
	count = map[string]int{}
	for i := 0; i < 3000; i++ {
		count[fake(t, f, "region[01].municipality.code")]++
	}
	if count["0180"] < 2500 || count["0184"] == 0 || len(count) != 2 {
		t.Fatalf("municipality draws inside 01 %v, want Stockholm ≈ 92%% and Solna the rest", count)
	}
}

func TestTableCellsAreStringNodes(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"w.json":     `"x"`,
		"place.json": `{"format":"{zip} {name}","rows":"place.tsv","key":"name"}`,
		"place.tsv":  "name\tzip\ttag\nStockholm\t1{digits(2)} {digits(2)}\t{/w}\nTranås\t573 {digits(2)}\tplain\n",
	})
	f := newGenerator(t, dir, WithSeed(1))
	zip := regexp.MustCompile(`^(1\d\d \d\d Stockholm|573 \d\d Tranås)$`)
	for i := 0; i < 50; i++ {
		if v := fake(t, f, "place"); !zip.MatchString(v) {
			t.Fatalf("place = %q, want the cell's tokens expanded", v)
		}
	}
	if v := fake(t, f, "place[Stockholm].tag"); v != "x" {
		t.Fatalf("place[Stockholm].tag = %q, want the reference rendered", v)
	}
	bad := writeFiles(t, map[string]string{
		"place.json": `{"format":"{name}","rows":"place.tsv"}`,
		"place.tsv":  "name\tzip\nA\t{name}\nB\t2\n",
	})
	if _, err := New(WithoutShippedData(), WithDataPath(bad)); err == nil || !strings.Contains(err.Error(), "place.tsv") || !strings.Contains(err.Error(), `no field "name"`) {
		t.Fatalf("New = %v, want a cell reading a column refused, naming the file", err)
	}
}

func TestTableSelectsARowByKeyOrName(t *testing.T) {
	files := with(geo(), map[string]string{
		"city.json": `{"format":"{name}","rows":"city.tsv","key":"code","name":"name"}`,
		"city.tsv":  "code\tname\nSTL\tSt. Louis\nSPI\tSpringfield\n",
	})
	f := newGenerator(t, writeFiles(t, files), WithSeed(1))
	for path, want := range map[string]string{
		"region[12]":                               "Skåne län",
		"region[Skåne län].code":                   "12",
		"municipality[1281].locality[Sandby]":      "Sandby",
		"municipality[1281].locality[Sandby].code": "L7",
		"municipality[0184].locality[Sandby].code": "L8",
		"city[St. Louis].code":                     "STL",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
	for path, want := range map[string]string{
		"region[99]":                    `"99"`,
		"locality[Sandby]":              "L7, L8",
		"region[12].code[1]":            "not a table",
		"region[]":                      "empty",
		"region[12":                     "]",
		"region[12].municipality[1480]": "not inside",
	} {
		if _, err := f.Fake(path); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Fake(%q) = %v, want an error mentioning %s", path, err, want)
		}
	}
	if v := fake(t, f, "municipality[0180].region"); v != "01" {
		t.Fatalf("municipality[0180].region = %q, want the link column's cell", v)
	}
	noKey := writeFiles(t, map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\nx\ny\n"})
	if _, err := newGenerator(t, noKey).Fake("t[x]"); err == nil || !strings.Contains(err.Error(), "no key") {
		t.Fatalf("Fake(t[x]) = %v, want no column to select by", err)
	}
}

func TestTableDescendsToALinkedTable(t *testing.T) {
	f := newGenerator(t, writeFiles(t, geo()), WithSeed(2))
	for i := 0; i < 200; i++ {
		if m := fake(t, f, "region[12].municipality.code"); regionOf[m] != "12" {
			t.Fatalf("region[12].municipality.code = %q, outside 12", m)
		}
		if l := fake(t, f, "region[12].locality.code"); regionOf[municipalityOf[l]] != "12" {
			t.Fatalf("region[12].locality.code = %q, outside 12", l)
		}
		if l := fake(t, f, "municipality[1480].locality"); l != "Göteborg" && l != "Hisingen" {
			t.Fatalf("municipality[1480].locality = %q", l)
		}
		if l := fake(t, f, "region.municipality.locality.name"); l == "" {
			t.Fatal("region.municipality.locality.name rendered empty")
		}
	}
	r, err := f.FakeRecord("region[14].municipality")
	if err != nil {
		t.Fatal(err)
	}
	if line := r.CSVLine(); line != "1480,Göteborg,590000,14" {
		t.Fatalf("record inside region 14 = %q", line)
	}
}

func TestLinkedTablesDrawConsistently(t *testing.T) {
	spellings := map[string]string{
		"leaf first":   `{"format":"{l}|{m}|{r}","l":"{/locality.code}","m":"{/municipality.code}","r":"{/region.code}"}`,
		"root first":   `{"format":"{r}|{m}|{l}","r":"{/region.code}","m":"{/municipality.code}","l":"{/locality.code}"}`,
		"skipping one": `{"format":"{r}|{l}","r":"{/region.code}","l":"{/locality.code}"}`,
		"nested":       `{"format":"{a}|{l}","a":{"format":"{r}|{m}","r":"{/region.code}","m":"{/municipality.code}"},"l":"{/locality.code}"}`,
		"a record":     `{"format":"","l":"{/locality.code}","m":"{/municipality.code}","r":"{/region.code}"}`,
	}
	for name, addr := range spellings {
		f := newGenerator(t, writeFiles(t, with(geo(), map[string]string{"addr.json": addr})), WithSeed(5))
		seen := map[string]bool{}
		for i := 0; i < 300; i++ {
			var parts []string
			if name == "a record" {
				r, err := f.FakeRecord("addr")
				if err != nil {
					t.Fatal(err)
				}
				for _, c := range r.Columns() {
					parts = append(parts, c.Value)
				}
			} else {
				parts = strings.Split(fake(t, f, "addr"), "|")
			}
			var l, m, r string
			for _, p := range parts {
				switch {
				case strings.HasPrefix(p, "L"):
					l = p
				case len(p) == 4:
					m = p
				default:
					r = p
				}
			}
			if m == "" {
				m = municipalityOf[l]
			}
			if municipalityOf[l] != m || regionOf[m] != r {
				t.Fatalf("%s: %v is not one consistent draw", name, parts)
			}
			seen[l] = true
		}
		if len(seen) < 4 {
			t.Fatalf("%s: only %v drawn in 300 renders", name, seen)
		}
	}
}

func TestLinkedTablesDrawApartAcrossGroupsAndRepeats(t *testing.T) {
	files := with(geo(), map[string]string{
		"groups.json": `{"format":"{a}|{b}","a":{"format":"{/locality.code}","drawGroup":"g"},"b":"{/locality.code}"}`,
		"many.json":   `{"format":"{x}","x":{"format":"{/locality.code}","repeat":8,"separator":"|"}}`,
	})
	f := newGenerator(t, writeFiles(t, files), WithSeed(9))
	for _, path := range []string{"groups", "many"} {
		differ := false
		for i := 0; i < 100 && !differ; i++ {
			parts := strings.Split(fake(t, f, path), "|")
			for _, p := range parts[1:] {
				differ = differ || p != parts[0]
			}
		}
		if !differ {
			t.Fatalf("%s: every draw agreed in 100 renders, want groups and repeat iterations drawn apart", path)
		}
	}
}

func TestSiblingTablesDrawInsideOneAncestor(t *testing.T) {
	files := with(siblings(), map[string]string{
		"addr.json": `{"format":"{s}|{p}|{l}","s":"{/street.name}","p":"{/postal-code.code}","l":"{/locality.code}"}`,
	})
	f := newGenerator(t, writeFiles(t, files), WithSeed(3))
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		parts := strings.Split(fake(t, f, "addr"), "|")
		if s, p, l := parts[0], parts[1], parts[2]; !slices.Contains(localityOfStreet[s], l) || localityOfCode[p] != l {
			t.Fatalf("addr = %v, want the street and the postal code inside the locality", parts)
		}
		seen[parts[2]] = true
	}
	if len(seen) < 4 {
		t.Fatalf("only %v drawn in 300 renders", seen)
	}
	for i := 0; i < 100; i++ {
		if s := fake(t, f, "locality[L4].street.name"); s != "Stora Södergatan" && s != "Klostergatan" {
			t.Fatalf("locality[L4].street.name = %q, outside L4", s)
		}
		if p := fake(t, f, "region[01].postal-code"); localityOfCode[p] != "L1" && localityOfCode[p] != "L2" && localityOfCode[p] != "L8" {
			t.Fatalf("region[01].postal-code = %q, outside region 01", p)
		}
	}
	if _, err := f.Fake("street[Avenyn]"); err == nil || !strings.Contains(err.Error(), "no key") {
		t.Fatalf("Fake(street[Avenyn]) = %v, want no column to select by", err)
	}
	paths := f.List()
	for _, p := range []string{"locality.postal-code", "locality.street", "region.municipality.locality.street.name"} {
		if !slices.Contains(paths, p) {
			t.Fatalf("List() lacks %s: %v", p, paths)
		}
	}
}

func TestTableSelectorInAReference(t *testing.T) {
	files := with(geo(), map[string]string{
		"x.json": `{"format":"{a} {b} {c}","a":"{/region[12].name}","b":"{/region[12].municipality.code}","c":"{/region[12].locality.code}"}`,
	})
	f := newGenerator(t, writeFiles(t, files), WithSeed(1))
	for i := 0; i < 100; i++ {
		parts := strings.Fields(fake(t, f, "x"))
		m, l := parts[len(parts)-2], parts[len(parts)-1]
		if !strings.HasPrefix(fake(t, f, "x"), "Skåne län") || regionOf[m] != "12" || municipalityOf[l] != m {
			t.Fatalf("x = %v, want one draw inside region 12", parts)
		}
	}
	for name, c := range map[string]struct{ json, want string }{
		"unknown row":          {`"{/region[99].name}"`, `"99"`},
		"ambiguous name":       {`"{/locality[Sandby].name}"`, "L7, L8"},
		"not inside":           {`"{/region[12].municipality[0180].name}"`, "not inside"},
		"selector on template": {`"{/x[1].a}"`, "not a table"},
	} {
		_, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, with(files, map[string]string{"bad.json": c.json}))))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: New = %v, want an error mentioning %q", name, err, c.want)
		}
	}
}

func TestTableSelectionFences(t *testing.T) {
	rejected := map[string]struct{ json, want string }{
		"bare beside a path":            {`"{/region} {/region.name}"`, "reads a path into"},
		"bare beside a linked path":     {`"{/region} {/locality.name}"`, "drawGroup"},
		"selected beside unselected":    {`"{/region[12].municipality.name} {/municipality.code}"`, "{/region[12].municipality.code}"},
		"two selections":                {`"{/region[12].name} {/region[14].name}"`, "drawGroup"},
		"selection beside a descendant": {`"{/region[12].name} {/locality.code}"`, "{/region[12].locality.code}"},
	}
	for name, c := range rejected {
		_, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, with(geo(), map[string]string{"x.json": c.json}))))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: New = %v, want it rejected mentioning %q", name, err, c.want)
		}
	}
	accepted := map[string]string{
		"a descent beside a path into it": `"{/region.municipality} {/region.municipality.name}"`,
		"selections in two groups":        `{"format":"{a} {b}","a":{"format":"{/region[12].name}","drawGroup":"a"},"b":"{/region[14].name}"}`,
		"one selection twice":             `"{/region[12].name} {/region[12].timezone} {/region[12].locality.name}"`,
		"a bare table in another group":   `{"format":"{a} {b}","a":{"format":"{/region}","drawGroup":"a"},"b":"{/locality.name}"}`,
	}
	for name, json := range accepted {
		f, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, with(geo(), map[string]string{"x.json": json}))), WithSeed(1))
		if err != nil {
			t.Errorf("%s: New = %v, want it accepted", name, err)
			continue
		}
		fake(t, f, "x")
	}
	f := newGenerator(t, writeFiles(t, with(geo(), map[string]string{"x.json": accepted["a descent beside a path into it"]})), WithSeed(1))
	for i := 0; i < 100; i++ {
		if parts := strings.Fields(fake(t, f, "x")); parts[0] != parts[1] {
			t.Fatalf("x = %v, want the descended row and the path into it to agree", parts)
		}
	}
}

// The family fence replays each read's selectors through the walk the render uses, so
// spellings that pin the same rows by different routes agree, and two rows of one
// table are refused at load rather than found at render.
func TestTableFamilyFenceReplaysPins(t *testing.T) {
	accepted := map[string]string{
		"skip-level selectors":       `"{/region[12].locality[L4].name}|{/region[12].name}"`,
		"descendant then ancestor":   `"{/locality[L4].name}|{/region[12].name}"`,
		"two selected levels":        `"{/municipality[1281].name}|{/locality[L4].name}"`,
		"nested prefixes":            `"{/region[12].municipality[1281].name}|{/municipality[1281].locality[L4].name}"`,
		"a draw under the selection": `"{/region[12].name}|{/region[12].municipality.locality[L4].name}"`,
	}
	for name, json := range accepted {
		f, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, with(geo(), map[string]string{"x.json": json}))), WithSeed(1))
		if err != nil {
			t.Errorf("%s: New = %v, want it accepted", name, err)
			continue
		}
		for i := 0; i < 50; i++ {
			v := fake(t, f, "x")
			if strings.Contains(v, "Sandby") || strings.Contains(v, "Malmö") || strings.Contains(v, "1280") || !strings.Contains(v, "Lund") && !strings.Contains(v, "Skåne") {
				t.Fatalf("%s: x = %q, want every part inside region 12 and Lund", name, v)
			}
		}
	}
	rejected := map[string]struct{ json, want string }{
		"two rows of one table":             {`"{/region[12].locality[L4].name} {/region[12].locality[L7].name}"`, "drawGroup"},
		"a row outside a selected ancestor": {`"{/region[14].name} {/locality[L4].name}"`, "drawGroup"},
	}
	for name, c := range rejected {
		_, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, with(geo(), map[string]string{"x.json": c.json}))))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: New = %v, want it rejected mentioning %q", name, err, c.want)
		}
	}
}

func TestTableFences(t *testing.T) {
	base := map[string]string{
		"region.json": `{"format":"{name}","rows":"region.tsv","key":"code","name":"name"}`,
		"region.tsv":  "code\tname\n01\tA\n12\tB\n",
	}
	rejected := map[string]struct {
		files map[string]string
		want  string
	}{
		"rows names a missing file":                      {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`}, "t.tsv"},
		"a TSV nothing names":                            {with(base, map[string]string{"stray.tsv": "a\nx\n"}), "stray.tsv"},
		"rows outside its folder":                        {with(base, map[string]string{"t.json": `{"format":"{code}","rows":"../region.tsv"}`}), "beside"},
		"rows not a tsv":                                 {with(base, map[string]string{"t.json": `{"format":"{code}","rows":"region.txt"}`, "region.txt": "code\n1\n"}), ".tsv"},
		"key names no column":                            {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"b"}`, "t.tsv": "a\nx\ny\n"}, `"b"`},
		"name names no column":                           {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","name":"b"}`, "t.tsv": "a\nx\ny\n"}, `"b"`},
		"weight names no column":                         {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","weight":"b"}`, "t.tsv": "a\nx\ny\n"}, `"b"`},
		"parent names no column":                         {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","parent":"b"}`, "t.tsv": "a\nx\ny\n"}, `"b"`},
		"key equals name":                                {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a","name":"a"}`, "t.tsv": "a\nx\ny\n"}, "drop"},
		"duplicate key":                                  {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a"}`, "t.tsv": "a\nx\nx\n"}, `"x"`},
		"empty key":                                      {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a"}`, "t.tsv": "a\tb\n\ty\nx\tz\n"}, "empty"},
		"a bracket in a key":                             {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a"}`, "t.tsv": "a\nx[1]\ny\n"}, `"["`},
		"a brace in a name":                              {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a","name":"n"}`, "t.tsv": "a\tn\nx\tx{1}\ny\ty\n"}, `"{"`},
		"name without a key":                             {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","name":"a"}`, "t.tsv": "a\nx\nx\n"}, "key"},
		"a name repeating inside one parent row":         {with(siblings(), map[string]string{"street.json": `{"format":"{name}","rows":"street.tsv","name":"name","parent":"locality"}`, "street.tsv": "name\tlocality\nAvenyn\tL1\nAvenyn\tL1\nStorgatan\tL2\nStorgatan\tL3\nStorgatan\tL4\nStorgatan\tL5\nStorgatan\tL6\nStorgatan\tL7\nStorgatan\tL8\n"}), `"Avenyn"`},
		"a name that is another row's key":               {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a","name":"n"}`, "t.tsv": "a\tn\nx\ty\ny\tz\n"}, `"y"`},
		"a cell reading its family":                      {with(geo(), map[string]string{"locality.tsv": "code\tname\tmunicipality\tnote\nL1\tStockholm\t0180\t{/municipality.code}\nL2\tSolna\t0184\t-\nL3\tMalmö\t1280\t-\nL4\tLund\t1281\t-\nL5\tGöteborg\t1480\t-\n"}), "family"},
		"a format reading its family":                    {with(geo(), map[string]string{"locality.json": `{"format":"{name} {/region.name}","rows":"locality.tsv","key":"code","name":"name","parent":"municipality"}`}), "family"},
		"a descendant named like an ancestor's column":   {with(geo(), map[string]string{"region.tsv": "code\tname\tpopulation\tlocality\n01\tStockholms län\t2400000\tx\n12\tSkåne län\t1400000\ty\n14\tVästra Götalands län\t1750000\tz\n"}), `"locality"`},
		"weight not a number":                            {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","weight":"w"}`, "t.tsv": "a\tw\nx\tmany\ny\t2\n"}, `"many"`},
		"weight zero":                                    {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","weight":"w"}`, "t.tsv": "a\tw\nx\t0\ny\t2\n"}, "0"},
		"weight negative":                                {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","weight":"w"}`, "t.tsv": "a\tw\nx\t-1\ny\t2\n"}, "-1"},
		"reserved column name":                           {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\tb.c\nx\ty\n"}, `"b.c"`},
		"duplicate column":                               {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\ta\nx\ty\n"}, `"a"`},
		"empty column name":                              {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\t\nx\ty\n"}, "empty"},
		"short row":                                      {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\tb\nx\ty\nz\n"}, "line 3"},
		"no rows":                                        {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\n"}, "no rows"},
		"a blank line after the rows":                    {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\tb\nx\t1\ny\t2\n\n"}, "line 4"},
		"a blank line in one column":                     {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\nx\n\ny\n"}, "line 3"},
		"a table reaching its family through a template": {with(geo(), map[string]string{"addr.json": `"{/locality.name}"`, "region.tsv": "code\tname\tpopulation\ttimezone\tnote\n01\tStockholms län\t2400000\tEurope/Stockholm\t{/addr}\n12\tSkåne län\t1400000\tEurope/Stockholm\t-\n14\tVästra Götalands län\t1750000\tEurope/Stockholm\t-\n"}), "family"},
		"a table reaching its family through a repeat":   {with(geo(), map[string]string{"addr.json": `{"format":"{/locality.name} ","repeat":3}`, "region.tsv": "code\tname\tpopulation\ttimezone\tnote\n01\tStockholms län\t2400000\tEurope/Stockholm\t{/addr}\n12\tSkåne län\t1400000\tEurope/Stockholm\t-\n14\tVästra Götalands län\t1750000\tEurope/Stockholm\t-\n"}), "family"},
		"a table reaching its family through a group":    {with(geo(), map[string]string{"addr.json": `{"format":"{/locality.name}","drawGroup":"g"}`, "region.tsv": "code\tname\tpopulation\ttimezone\tnote\n01\tStockholms län\t2400000\tEurope/Stockholm\t{/addr}\n12\tSkåne län\t1400000\tEurope/Stockholm\t-\n14\tVästra Götalands län\t1750000\tEurope/Stockholm\t-\n"}), "family"},
		"one row":                       {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\nx\n"}, "one row"},
		"empty file":                    {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": ""}, "header"},
		"parent is not a table":         {map[string]string{"p.json": `"x"`, "t.json": `{"format":"{a}","rows":"t.tsv","parent":"p"}`, "t.tsv": "a\tp\nx\tx\ny\tx\n"}, "not a table"},
		"parent does not exist":         {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","parent":"p"}`, "t.tsv": "a\tp\nx\tx\ny\tx\n"}, `"p"`},
		"parent in another folder":      {map[string]string{"g/p.json": `{"format":"{k}","rows":"p.tsv","key":"k"}`, "g/p.tsv": "k\nx\ny\n", "t.json": `{"format":"{a}","rows":"t.tsv","parent":"p"}`, "t.tsv": "a\tp\nx\tx\ny\ty\n"}, `"p"`},
		"parent has no key":             {map[string]string{"p.json": `{"format":"{k}","rows":"p.tsv"}`, "p.tsv": "k\nx\ny\n", "t.json": `{"format":"{a}","rows":"t.tsv","parent":"p"}`, "t.tsv": "a\tp\nx\tx\ny\ty\n"}, "key"},
		"dangling link":                 {with(base, map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","parent":"region"}`, "t.tsv": "a\tregion\nx\t01\ny\t99\n"}), `"99"`},
		"childless parent row":          {with(base, map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","parent":"region"}`, "t.tsv": "a\tregion\nx\t01\ny\t01\n"}), `"12"`},
		"parent cycle":                  {map[string]string{"a.json": `{"format":"{k}","rows":"a.tsv","key":"k","parent":"b"}`, "a.tsv": "k\tb\nx\tx\ny\ty\n", "b.json": `{"format":"{k}","rows":"b.tsv","key":"k","parent":"a"}`, "b.tsv": "k\ta\nx\tx\ny\ty\n"}, "cycle"},
		"child named like a column":     {with(base, map[string]string{"name.json": `{"format":"{a}","rows":"name.tsv","parent":"region"}`, "name.tsv": "a\tregion\nx\t01\ny\t12\n"}), `"name"`},
		"rows nested in a field":        {map[string]string{"t.json": `{"format":"{x}","x":{"format":"{a}","rows":"x.tsv"}}`, "x.tsv": "a\nx\ny\n"}, "category"},
		"rows in a choice item":         {map[string]string{"t.json": `[{"format":"{a}","rows":"t.tsv"},"y"]`, "t.tsv": "a\nx\ny\n"}, "category"},
		"unknown table option":          {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","fields":"a"}`, "t.tsv": "a\nx\ny\n"}, "a table takes"},
		"repeat on a table":             {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","repeat":2}`, "t.tsv": "a\nx\ny\n"}, "a table takes"},
		"drawGroup on a table":          {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","drawGroup":"g"}`, "t.tsv": "a\nx\ny\n"}, "a table takes"},
		"option not a string":           {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":1}`, "t.tsv": "a\nx\ny\n"}, "string"},
		"format names no column":        {map[string]string{"t.json": `{"format":"{b}","rows":"t.tsv"}`, "t.tsv": "a\nx\ny\n"}, `no column "b"`},
		"format reads into a column":    {map[string]string{"t.json": `{"format":"{a.x}","rows":"t.tsv"}`, "t.tsv": "a\nx\ny\n"}, `"a"`},
		"a reference into a column":     {with(base, map[string]string{"t.json": `"{/region.name.x}"`}), "column"},
		"a category referencing itself": {map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv"}`, "t.tsv": "a\n{/t.a}\ny\n"}, "names the category it sits in"},
	}
	for name, c := range rejected {
		_, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, c.files)))
		if err == nil {
			t.Errorf("%s: New = nil error, want it rejected at load", name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: New = %v, want it to mention %q", name, err, c.want)
		}
	}
	if _, err := New(WithoutShippedData(), WithDataFS(os.DirFS(writeFiles(t, base)))); err != nil {
		t.Fatalf("New(WithDataFS) = %v, want a table loaded from any fs.FS", err)
	}
	inline := newGenerator(t, writeFiles(t, base))
	if _, err := inline.NewTemplate(`{"format":"{a}","rows":"t.tsv"}`); err == nil || !strings.Contains(err.Error(), "inline") {
		t.Fatalf("NewTemplate(rows) = %v, want a table refused inline", err)
	}
	bom := newGenerator(t, writeFiles(t, map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a"}`, "t.tsv": "\xEF\xBB\xBFa\tb\nx\t1\ny\t2\n"}), WithSeed(1))
	if v := fake(t, bom, "t[x].b"); v != "1" {
		t.Fatalf("t[x].b under a BOM header = %q, want the mark stripped", v)
	}
	crlf := newGenerator(t, writeFiles(t, map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":"a"}`, "t.tsv": "a\tb\r\nx\t1\r\ny\t2\r\n"}), WithSeed(1))
	if v := fake(t, crlf, "t[y].b"); v != "2" {
		t.Fatalf("t[y].b in a CRLF file = %q, want the carriage return stripped", v)
	}
	a, b := newGenerator(t, writeFiles(t, geo()), WithSeed(1)), newGenerator(t, writeFiles(t, geo()), WithSeed(1))
	for _, path := range []string{"region.nope", "region.locality.nope", "region.municipality[9999]", "region[12].locality[L1]"} {
		if _, err := a.Fake(path); err == nil {
			t.Fatalf("Fake(%s) = nil error", path)
		}
		if _, err := a.FakeRecord(path); err == nil {
			t.Fatalf("FakeRecord(%s) = nil error", path)
		}
	}
	if x, y := fake(t, a, "region.locality.name"), fake(t, b, "region.locality.name"); x != y {
		t.Fatalf("a failed Fake shifted the seeded stream: %q != %q", x, y)
	}
	two := newGenerator(t, writeFiles(t, with(geo(), map[string]string{"capital.json": `{"format":"{timezone}","rows":"region.tsv","key":"code"}`})), WithSeed(1))
	if v := fake(t, two, "capital[12]"); v != "Europe/Stockholm" {
		t.Fatalf("capital[12] over region.tsv = %q, want a second category over one file", v)
	}
	empty := newGenerator(t, writeFiles(t, map[string]string{"t.json": `{"format":"","rows":"t.tsv","key":"a"}`, "t.tsv": "a\tb\nx\t1\ny\t2\n"}), WithSeed(1))
	if v := fake(t, empty, "t"); v != "" {
		t.Fatalf("a record-only table renders %q, want \"\"", v)
	}
	if v := fake(t, empty, "t[y].b"); v != "2" {
		t.Fatalf("t[y].b = %q", v)
	}
}

func TestSameShapedChoiceIsATable(t *testing.T) {
	rows := `[{"format":"{name}","name":"Sweden","alpha2":"SE"},{"format":"{name}","name":"Norway","alpha2":"NO"},{"format":"{name}","name":"Denmark","alpha2":"DK"}]`
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"country": rows})))
	for _, want := range []string{"country.tsv", `"rows"`, `alpha2\tname`, "3 "} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("New(same-shaped choice) = %v, want it refused mentioning %q", err, want)
		}
	}
	weighted := `[{"format":"{name}","name":"Sweden","weight":2},{"format":"{name}","name":"Norway"}]`
	if _, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"country": weighted}))); err == nil || !strings.Contains(err.Error(), "weight") {
		t.Fatalf("New(weighted same-shaped choice) = %v, want it refused naming a weight column", err)
	}
	accepted := map[string]string{
		"nested":            `{"format":"{c}","c":` + rows + `}`,
		"a choice field":    `[{"format":"{maker} {model}","maker":"BMW","model":["X3","X5"]},{"format":"{maker} {model}","maker":"Ford","model":["Focus","Fiesta"]}]`,
		"different formats": `[{"format":"{a}-{b}","a":"1","b":"2"},{"format":"{b}-{a}","a":"3","b":"4"}]`,
		"different fields":  `[{"format":"{a}","a":"1","b":"2"},{"format":"{a}","a":"3"}]`,
		"a string item":     `[{"format":"{a}","a":"1"},"x"]`,
		"strings":           `["a","b","c"]`,
	}
	for name, json := range accepted {
		if _, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"x": json}))); err != nil {
			t.Errorf("%s: New = %v, want it accepted", name, err)
		}
	}
	f := newGenerator(t, writeData(t, map[string]string{"w": `"x"`}))
	if _, err := f.NewTemplate(rows); err != nil {
		t.Fatalf("NewTemplate(same-shaped choice) = %v, want an inline template exempt", err)
	}
}

func TestTableInAStructTag(t *testing.T) {
	f := newGenerator(t, writeFiles(t, geo()), WithSeed(1))
	var v struct {
		Region   string `fake:"region[12].name"`
		Locality string `fake:"region[12].locality.code"`
		Any      string `fake:"{/region[Skåne län].timezone}-x"` // the row region[12] names, by name
	}
	if err := f.FakeStruct(&v); err != nil {
		t.Fatal(err)
	}
	if v.Region != "Skåne län" || regionOf[municipalityOf[v.Locality]] != "12" || v.Any != "Europe/Stockholm-x" {
		t.Fatalf("FakeStruct = %+v, want the selected rows", v)
	}
}

func TestTableRowsAreAlternatives(t *testing.T) {
	files := map[string]string{
		"cur.json":     `{"format":"{code}","rows":"cur.tsv","key":"code"}`,
		"cur.tsv":      "code\tsym\nEUR\t€\nSEK\tkr\nUSD\t$\n",
		"y.json":       `{"format":"{sym}","rows":"y.tsv","key":"k"}`,
		"y.tsv":        "k\tsym\n1\t{/cur[SEK].sym}\n2\t{/cur[USD].sym}\n",
		"country.json": `{"format":"{name} {money}","rows":"country.tsv","key":"alpha2"}`,
		"country.tsv":  "alpha2\tname\tmoney\nFI\tFinland\t{/cur[EUR].sym}\nSE\tSweden\t{/cur[SEK].sym}\n",
	}
	f := newGenerator(t, writeFiles(t, files), WithSeed(1))
	if v := fake(t, f, "country[SE]"); v != "Sweden kr" {
		t.Fatalf("country[SE] = %q", v)
	}
	if v := fake(t, f, "country[FI].money"); v != "€" {
		t.Fatalf("country[FI].money = %q", v)
	}
	accepted := map[string]map[string]string{
		"a cell reaching a table whose rows are alternatives": {"country.tsv": "alpha2\tname\tmoney\nFI\tFinland\t{/y.sym}\nSE\tSweden\t{/cur[SEK].sym}\n"},
		"a cell agreeing with the row it selects":             {"country.tsv": "alpha2\tname\tmoney\nFI\tFinland\t{/y[1].sym} {/cur[SEK].sym}\nSE\tSweden\t{/cur[SEK].sym}\n"},
		"a selected row's cell beside the row it agrees with": {"z.json": `"{/country[SE].money} {/cur[SEK].sym}"`},
		"a selected row whole beside the row it agrees with":  {"z.json": `"{/country[SE]} {/cur[SEK].sym}"`},
		"a column under a selected ancestor": {
			"city.json": `{"format":"{name} {money}","rows":"city.tsv","key":"name","parent":"country"}`,
			"city.tsv":  "name\tcountry\tmoney\nHelsinki\tFI\t{/cur[EUR].sym}\nMalmö\tSE\t{/cur[SEK].sym}\nLund\tSE\t{/cur[SEK].sym}\n",
			"z.json":    `"{/country[SE].city.money} {/cur[SEK].sym}"`,
		},
	}
	symbols := regexp.MustCompile(`€|kr|\$`)
	for name, more := range accepted {
		g, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, with(files, more))), WithSeed(1))
		if err != nil {
			t.Errorf("%s: New = %v, want it accepted", name, err)
			continue
		}
		paths := []string{"country", "country[FI]"}
		if _, defined := more["z.json"]; defined {
			paths = append(paths, "z")
		}
		for i := 0; i < 50; i++ {
			for _, path := range paths {
				v := fake(t, g, path)
				seen := map[string]bool{}
				for _, s := range symbols.FindAllString(v, -1) {
					seen[s] = true
				}
				if len(seen) > 1 {
					t.Fatalf("%s: %s = %q, want one currency per render", name, path, v)
				}
			}
		}
	}
	rejected := map[string]map[string]string{
		"one cell selecting two rows":     {"country.tsv": "alpha2\tname\tmoney\nFI\tFinland\t{/cur[EUR].sym}{/cur[SEK].sym}\nSE\tSweden\t{/cur[SEK].sym}\n"},
		"two tables' cells in one render": {"z.json": `"{/country.money} {/y.sym}"`},
		"the format beside a cell":        {"country.json": `{"format":"{money} {/cur[SEK].sym}","rows":"country.tsv","key":"alpha2"}`},
		"two columns of one row":          {"country.json": `{"format":"{a} {b}","rows":"country.tsv","key":"alpha2"}`, "country.tsv": "alpha2\ta\tb\nFI\t{/cur[SEK].sym}\t{/cur[EUR].sym}\nSE\t{/cur[EUR].sym}\t{/cur[SEK].sym}\n"},
		"a whole read beside a path":      {"country.json": `{"format":"{a} {b}","rows":"country.tsv","key":"alpha2"}`, "country.tsv": "alpha2\ta\tb\nFI\t{/cur}\t{/cur[EUR].sym}\nSE\t{/cur}\t{/cur[SEK].sym}\n"},
		"a cell reaching another's cells": {"y.json": `{"format":"{sym}","rows":"y.tsv","key":"k"}`, "y.tsv": "k\tsym\n1\t{/cur[SEK].sym}\n2\t{/cur[USD].sym}\n", "cur.tsv": "code\tsym\nEUR\t€\nSEK\tkr\nUSD\t$\n", "country.tsv": "alpha2\tname\tmoney\nFI\tFinland\t{/y[1].sym} {/cur[EUR].sym}\nSE\tSweden\t{/cur[SEK].sym}\n"},
	}
	for name, more := range rejected {
		_, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, with(files, more))))
		if err == nil || !strings.Contains(err.Error(), "drawGroup") {
			t.Errorf("%s: New = %v, want it refused", name, err)
		}
	}
}

func TestTableOverridesByLayering(t *testing.T) {
	mine := writeFiles(t, map[string]string{
		"misc/language.json": `{"format":"{name}","rows":"language.tsv","key":"code"}`,
		"misc/language.tsv":  "code\tname\nxx\tNowhere\nyy\tElsewhere\n",
	})
	f, err := New(WithDataPath(mine), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	if v := fake(t, f, "misc.language[xx]"); v != "Nowhere" {
		t.Fatalf("misc.language[xx] = %q, want the layered table", v)
	}
	if paths := f.List(); slices.Contains(paths, "misc.language.code3") {
		t.Fatal("List() still offers the shipped table's columns under an overridden category")
	}
}

func TestShippedTables(t *testing.T) {
	f := newGenerator(t, "data/misc", WithSeed(1))
	for path, want := range map[string]string{
		"territory[SE]":                         "Sweden",
		"territory[Sweden].alpha3":              "SWE",
		"territory[SE].numeric":                 "752",
		"territory[SE].tld":                     ".se",
		"territory[SE].calling-code":            "46",
		"territory[SE].capital":                 "Stockholm",
		"territory[SE].currency":                "SEK",
		"territory[SE].flag":                    "🇸🇪",
		"territory[SE].country":                 "SE",
		"territory[GL]":                         "Greenland",
		"territory[GL].country":                 "DK",
		"territory[Åland Islands].country":      "FI",
		"currency[SEK].name":                    "Swedish Krona",
		"currency[SEK].symbol":                  "kr",
		"currency[SEK].numeric":                 "752",
		"currency[SEK].decimals":                "2",
		"currency[Euro].code":                   "EUR",
		"language[sv]":                          "Swedish",
		"language[sv].code3":                    "swe",
		"language[Swedish].code":                "sv",
		"language[nl]":                          "Dutch",
		"httpstatus[200]":                       "200 OK",
		"httpstatus[404]":                       "404 Not Found",
		"httpstatus[404].reason":                "Not Found",
		"httpstatus[451].reason":                "Unavailable For Legal Reasons",
		"httpstatus[500].reason":                "Internal Server Error",
		"mimetype[application/json].ext":        ".json",
		"mimetype[text/markdown].ext":           ".md",
		"mimetype[.jpg]":                        "image/jpeg",
		"mimetype[.mov]":                        "video/quicktime",
		"mimetype[.mp3]":                        "audio/mpeg",
		"mimetype[.mp4]":                        "video/mp4",
		"mimetype[.ogg]":                        "audio/ogg",
		"timezone[Europe/Stockholm]":            "Europe/Stockholm",
		"timezone[Europe/Stockholm].offset":     "+01:00",
		"timezone[Asia/Kathmandu].offset":       "+05:45",
		"timezone[America/New_York].offset":     "-05:00",
		"territory[SE].timezone":                "Europe/Stockholm",
		"httpmethod[GET].safe":                  "true",
		"httpmethod[GET].idempotent":            "true",
		"httpmethod[POST].safe":                 "false",
		"httpmethod[POST].idempotent":           "false",
		"httpmethod[PUT].idempotent":            "true",
		"protocol[TCP].number":                  "6",
		"protocol[UDP].number":                  "17",
		"protocol[ICMP].name":                   "Internet Control Message",
		"port[443]":                             "443",
		"port[443].service":                     "https",
		"port[22].service":                      "ssh",
		"port[3306].service":                    "mysql",
		"port[465].service":                     "urd",
		"port[2049].service":                    "nfs",
		"protocol[Transmission Control].number": "6",
		"tld[.se]":                              ".se",
		"tld[.se].type":                         "country-code",
		"tld[.com].type":                        "generic",
		"tld[.museum].type":                     "sponsored",
		"tld[.arpa].type":                       "infrastructure",
		"tld[.рф]":                              ".xn--p1ai",
		"tld[.xn--p1ai].unicode":                ".рф",
		"territory[NP].timezone.offset":         "+05:45",
		"loglevel[0]":                           "emerg",
		"loglevel[1]":                           "alert",
		"loglevel[2]":                           "crit",
		"loglevel[3]":                           "err",
		"loglevel[4]":                           "warning",
		"loglevel[5]":                           "notice",
		"loglevel[6]":                           "info",
		"loglevel[7]":                           "debug",
		"loglevel[err].code":                    "3",
		"loglevel[6].severity":                  "Informational",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
	re := map[string]*regexp.Regexp{
		"territory.numeric":      regexp.MustCompile(`^\d{3}$`),
		"territory.tld":          regexp.MustCompile(`^\.[a-z]{2}$`),
		"territory.calling-code": regexp.MustCompile(`^\d{1,4}(-\d{3})?$`),
		"territory.capital":      regexp.MustCompile(`\p{L}`),
		"territory.country":      regexp.MustCompile(`^[A-Z]{2}$`),
		"territory.currency":     regexp.MustCompile(`^[A-Z]{3}$`),
		"territory.flag":         regexp.MustCompile(`^[\x{1F1E6}-\x{1F1FF}]{2}$`),
		"territory.languages":    regexp.MustCompile(`^[a-z]{2,3}(-[A-Z]{2})?(,[a-z]{2,3}(-[A-Z]{2})?)*$`),
		"currency.numeric":       regexp.MustCompile(`^\d{3}$`),
		"currency.decimals":      regexp.MustCompile(`^[0-4]$`),
	}
	for i := 0; i < 100; i++ {
		for p, rx := range re {
			if v := fake(t, f, p); !rx.MatchString(v) {
				t.Fatalf("misc %s = %q, want %s", p, v, rx)
			}
		}
	}
	count := map[string]bool{}
	for i := 0; i < 5000; i++ {
		count[fake(t, f, "territory.alpha2")] = true
	}
	if len(count) < 200 {
		t.Fatalf("territory draws %d distinct rows in 5000, want the full register", len(count))
	}
	for i := 0; i < 200; i++ {
		got := fakeTemplate(t, f, `{/territory.alpha2} {/timezone.territory}`)
		if drawn, zone, _ := strings.Cut(got, " "); drawn != zone {
			t.Fatalf("%q: a territory and a timezone in one render disagree on the territory", got)
		}
	}
	var row struct {
		Idempotent bool   `fake:"misc.httpmethod.idempotent"`
		Method     string `fake:"misc.httpmethod.method"`
		Safe       bool   `fake:"misc.httpmethod.safe"`
	}
	if err := newGenerator(t, "data", WithSeed(1)).FakeStruct(&row); err != nil {
		t.Fatalf("FakeStruct into a bool = %v, want the register's yes and no shipped as true and false", err)
	}
	want, named := map[string][2]bool{
		"CONNECT": {false, false}, "DELETE": {false, true}, "GET": {true, true},
		"HEAD": {true, true}, "OPTIONS": {true, true}, "PATCH": {false, false},
		"POST": {false, false}, "PUT": {false, true}, "TRACE": {true, true},
	}[row.Method]
	if !named || row.Safe != want[0] || row.Idempotent != want[1] {
		t.Fatalf("%s drew safe=%v idempotent=%v", row.Method, row.Safe, row.Idempotent)
	}
	zones := map[string]int{}
	for i := 0; i < 2000; i++ {
		zones[fake(t, f, "territory[US].timezone")]++
	}
	if peopled := zones["America/New_York"] + zones["America/Chicago"] + zones["America/Los_Angeles"] + zones["America/Denver"]; peopled < 1600 {
		t.Fatalf("the four zones most Americans live in take %d of 2000 US draws, want the population weight to favour them", peopled)
	}
}

// TestEveryTerritoryNamesAShippedCountry proves what no parent can: the sovereign a
// territory names is a row of the same table, which a link would make a cycle.
func TestEveryTerritoryNamesAShippedCountry(t *testing.T) {
	head, rows := shippedRows(t, "territory.tsv", "alpha2", "country")
	keys := map[string]bool{}
	for _, r := range rows {
		keys[r[head["alpha2"]]] = true
	}
	for i, r := range rows {
		if c := r[head["country"]]; !keys[c] {
			t.Errorf("territory.tsv line %d: country %q is no alpha2 of the table", i+2, c)
		}
	}
}

// TestEveryUserAgentColumnAgreesWithItsString pins what the column regexes cannot:
// every Chromium fork carries Chrome's token, so a mis-ordered pattern relabels a
// row rather than dropping it.
func TestEveryUserAgentColumnAgreesWithItsString(t *testing.T) {
	head, rows := shippedRows(t, "useragent.tsv", "browser", "device", "os", "ua")
	for _, token := range []struct{ in, browser string }{
		{"Edg", "Edge"}, {"OPR/", "Opera"}, {"SamsungBrowser/", "Samsung Internet"}, {"CriOS/", "Chrome"},
	} {
		for i, r := range rows {
			if strings.Contains(r[head["ua"]], token.in) && r[head["browser"]] != token.browser {
				t.Errorf("useragent.tsv line %d: %q carries %q but is labelled %q", i+2, r[head["ua"]], token.in, r[head["browser"]])
			}
		}
	}
	for i, r := range rows {
		if r[head["device"]] == "mobile" && r[head["os"]] != "Android" && r[head["os"]] != "iOS" {
			t.Errorf("useragent.tsv line %d: a mobile row runs %q", i+2, r[head["os"]])
		}
	}
}

// TestEveryTimezoneRowIsWellFormed scans the file rather than drawing: the draw is
// population-weighted, so it samples the populous head and leaves most rows unrendered.
func TestEveryTimezoneRowIsWellFormed(t *testing.T) {
	head, rows := shippedRows(t, "timezone.tsv", "offset", "territory", "weight", "zone")
	offset := regexp.MustCompile(`^[+-](0\d|1[0-4]):[0-5]\d$`)
	for i, r := range rows {
		if !offset.MatchString(r[head["offset"]]) {
			t.Errorf("timezone.tsv line %d: offset %q is no ±HH:MM", i+2, r[head["offset"]])
		}
		if w, err := strconv.Atoi(r[head["weight"]]); err != nil || w < 1 {
			t.Errorf("timezone.tsv line %d: weight %q is not a positive number", i+2, r[head["weight"]])
		}
	}
}

// TestEveryTldRowSpellsItsKey scans the file rather than drawing: 100 draws leave
// most of 1438 rows unrendered, and the register wraps a right-to-left label in bidi
// marks that belong to its display, not to the label.
func TestEveryTldRowSpellsItsKey(t *testing.T) {
	head, rows := shippedRows(t, "tld.tsv", "tld", "type", "unicode")
	types := map[string]bool{"country-code": true, "generic": true, "generic-restricted": true, "infrastructure": true, "sponsored": true}
	for i, r := range rows {
		key, kind, shown := r[head["tld"]], r[head["type"]], r[head["unicode"]]
		if !types[kind] {
			t.Errorf("tld.tsv line %d: %s is typed %q, which the register assigns no TLD", i+2, key, kind)
		}
		for _, c := range shown {
			if unicode.Is(unicode.Cf, c) {
				t.Errorf("tld.tsv line %d: %s displays %q, carrying the format character %U", i+2, key, shown, c)
			}
		}
		if idn := strings.HasPrefix(key, ".xn--"); idn && shown == key {
			t.Errorf("tld.tsv line %d: the IDN %s displays its A-label, not a Unicode form", i+2, key)
		} else if !idn && shown != key {
			t.Errorf("tld.tsv line %d: the ASCII TLD %s displays %q", i+2, key, shown)
		}
	}
}

// shippedRows reads a shipped misc TSV as the loader does, proving it names the
// columns asked for and that every row fills them.
func shippedRows(t *testing.T, file string, want ...string) (map[string]int, [][]string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("data", "misc", file))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	head := map[string]int{}
	for i, name := range strings.Split(lines[0], "\t") {
		head[name] = i
	}
	for _, name := range want {
		if _, ok := head[name]; !ok {
			t.Fatalf("%s columns %v, want %v", file, lines[0], want)
		}
	}
	rows := make([][]string, 0, len(lines)-1)
	for i, line := range lines[1:] {
		cells := strings.Split(line, "\t")
		if len(cells) != len(head) {
			t.Fatalf("%s line %d has %d cells, want %d", file, i+2, len(cells), len(head))
		}
		rows = append(rows, cells)
	}
	return head, rows
}

// TestNamedTableWithoutAKeyResolvesInsideItsParent pins a table whose rows are
// told apart only by their parent: a name selects a row inside the parent pinned
// before it, an ambiguous one is listed by its parent's spelling, and a name
// repeating inside one parent row is a load error (see TestTableFences).
func TestNamedTableWithoutAKeyResolvesInsideItsParent(t *testing.T) {
	files := siblings()
	files["street.json"] = `{"format":"{name}","rows":"street.tsv","name":"name","parent":"locality","weight":"segments"}`
	f := newGenerator(t, writeFiles(t, files), WithSeed(1))
	for path, want := range map[string]string{
		"street[Avenyn]":                                                   "Avenyn",
		"street[Avenyn].locality":                                          "L5",
		"locality[L7].street[Sandbyvägen]":                                 "Sandbyvägen",
		"locality[L7].street[Sandbyvägen].segments":                        "2",
		"municipality[0184].locality[Sandby].street[Sandbyvägen]":          "Sandbyvägen",
		"municipality[0184].locality[Sandby].street[Sandbyvägen].locality": "L8",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
	for path, want := range map[string]string{
		"street[Sandbyvägen]":         "locality[L7].street[Sandbyvägen]",
		"locality[L1].street[Avenyn]": "not inside",
		"street[Kungsgatan]":          `"Kungsgatan"`,
	} {
		if _, err := f.Fake(path); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Fake(%q) = %v, want an error mentioning %s", path, err, want)
		}
	}
	if got := fakeTemplate(t, f, `{/locality[L8].name}: {/locality[L8].street[Sandbyvägen].name}`); got != "Sandby: Sandbyvägen" {
		t.Fatalf("a name selected inside a pinned parent = %q", got)
	}
}

// TestAmbiguousNameNamesARunnablePath pins what a shell user reads: each row is
// named as the path they can type, from the root, and the rows are listed plainly.
func TestAmbiguousNameNamesARunnablePath(t *testing.T) {
	files := map[string]string{}
	for name, body := range siblings() {
		files["se/"+name] = body
	}
	files["se/street.json"] = `{"format":"{name}","rows":"street.tsv","name":"name","parent":"locality","weight":"segments"}`
	f := newGenerator(t, writeFiles(t, files), WithSeed(1))
	_, err := f.Fake("se.street[Sandbyvägen]")
	if err == nil {
		t.Fatal("se.street[Sandbyvägen] resolved, though two rows carry that name")
	}
	for _, want := range []string{"se.locality[L7].street[Sandbyvägen]", "se.locality[L8].street[Sandbyvägen]"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%v does not name %s", err, want)
		}
		if got := fake(t, f, want); got != "Sandbyvägen" {
			t.Errorf("Fake(%q) = %q, want the named path to run", want, got)
		}
	}
	if strings.Contains(err.Error(), "[se.locality") {
		t.Errorf("%v lists the rows as a Go slice; separate them with commas", err)
	}
	if _, err := f.Fake("se.locality[Sandby]"); err == nil || !strings.Contains(err.Error(), "one of L7, L8") {
		t.Fatalf("se.locality[Sandby] = %v, want its keys listed plainly", err)
	}
}

// TestEnteredRowsAgreeWithPinning holds what a fence walk decides about the rows it entered, and
// the rows a render draws, to what pinning refuses: drift between them is data that loads and then
// renders a family that disagrees.
func TestEnteredRowsAgreeWithPinning(t *testing.T) {
	f := newGenerator(t, writeFiles(t, geo()), WithSeed(1))
	var tables []*table
	for _, category := range []string{"region", "municipality", "locality"} {
		tbl, isTable := f.categories[category].(*table)
		if !isTable {
			t.Fatalf("%s is not a table", category)
		}
		tables = append(tables, tbl)
	}
	var none pinSet
	for _, entered := range tables {
		for row := 0; row < entered.rows(); row++ {
			var drawn []tablePin
			alone := none.entered(entered, row, false)
			alone.each(func(tbl *table, r int) { drawn = append(drawn, tablePin{tbl, r}) })
			if len(drawn) != 1 || drawn[0] != (tablePin{entered, row}) {
				t.Errorf("a drawn %s enters %v, want its own row alone", entered.selectorSpelling(row), drawn)
			}
			s := none.entered(entered, row, true)
			for _, cand := range tables {
				for i := 0; i < 20; i++ {
					drawing := s.clone()
					if r := drawing.rowOf(f.rand, cand); s.clash(cand, r) != nil {
						t.Errorf("rowOf draws %s beside %s, which pinning it there refuses", cand.selectorSpelling(r), entered.selectorSpelling(row))
					}
				}
				for cr := 0; cr < cand.rows(); cr++ {
					beside := s.clone()
					refused := beside.pinRow(cand, cr) != nil
					if got := alternatives(drawAt{alt: s}, drawAt{alt: none.entered(cand, cr, true)}); got != refused {
						t.Errorf("alternatives(%s, %s) = %v, but pinning both refuses = %v", entered.selectorSpelling(row), cand.selectorSpelling(cr), got, refused)
					}
				}
			}
		}
	}
}

// TestCellReadsMeetWhereTheRenderPairsTheRows reads what the TSVs below do not show: region 01's
// note holds {/addr} and 1280's, under region 12, holds {/addr.city}, so the two meet only where
// no read pins a region.
func TestCellReadsMeetWhereTheRenderPairsTheRows(t *testing.T) {
	cells := with(geo(), map[string]string{
		"addr.json":        `{"format":"{city}","city":["Lund","Malmö"]}`,
		"municipality.tsv": "code\tname\tregion\tpopulation\tnote\n0180\tStockholm\t01\t980000\t-\n0184\tSolna\t01\t85000\t-\n1280\tMalmö\t12\t360000\t{/addr.city}\n1281\tLund\t12\t130000\t-\n1480\tGöteborg\t14\t590000\t-\n",
		"region.tsv":       "code\tname\tpopulation\tnote\n01\tStockholms län\t2400000\t{/addr}\n12\tSkåne län\t1400000\t-\n14\tVästra Götalands län\t1750000\t-\n",
	})
	paths := with(cells, map[string]string{"x.json": `"{/region.note} {/municipality.note}"`})
	f, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, paths)), WithSeed(1))
	if err != nil {
		t.Fatalf("New = %v, want the two cells accepted", err)
	}
	fake(t, f, "x")
	whole := with(cells, map[string]string{
		"municipality.json": `{"format":"{name}={note}","rows":"municipality.tsv","key":"code","name":"name","parent":"region","weight":"population"}`,
		"region.json":       `{"format":"{name}={note}","rows":"region.tsv","key":"code","name":"name","weight":"population"}`,
		"x.json":            `"{/region}|{/municipality}"`,
	})
	if _, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, whole)), WithSeed(1)); err == nil || !strings.Contains(err.Error(), "reads a path into") {
		t.Fatalf("New = %v, want the two drawn rows held to one draw of addr", err)
	}
}
