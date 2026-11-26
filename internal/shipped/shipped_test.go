package shipped

import (
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func module(namespace string, files map[string]string) fs.FS {
	out := fstest.MapFS{}
	if namespace != "" {
		out[".fejkdata.json"] = &fstest.MapFile{Data: []byte(`{"namespace":"` + namespace + `"}`)}
	}
	for name, body := range files {
		out[name] = &fstest.MapFile{Data: []byte(body)}
	}
	return out
}

func loaded(t *testing.T, modules ...fs.FS) *Data {
	t.Helper()
	d, err := Load(modules)
	if err != nil {
		t.Fatalf("Load = %v", err)
	}
	return d
}

func colorsAndSizes(t *testing.T) *Data {
	t.Helper()
	return loaded(t, module("ns", map[string]string{
		"color.json": `["red"]`,
		"size.json":  `{"format":"{n}","n":["1"]}`,
	}))
}

func TestLoadRefusesAModuleUnderNoNamespace(t *testing.T) {
	_, err := Load([]fs.FS{module("ns", map[string]string{"a.json": `"x"`}), module("", map[string]string{"b.json": `"y"`})})
	if err == nil || !strings.Contains(err.Error(), "module 1 names no namespace") {
		t.Fatalf("Load = %v, want module 1 refused for naming no namespace", err)
	}
}

func TestListingGivesAPathItsOwnRowElseItsCategorys(t *testing.T) {
	d := colorsAndSizes(t)
	got, err := d.Listing([]byte("ns.color\treal, valid\tthe palette\nns.size\tinvented\tmade up\n"), ListingState)
	if err != nil {
		t.Fatal(err)
	}
	want := []Path{
		{"ns.color", []string{"real", "valid"}, "red", "the palette"},
		{"ns.size", []string{"invented"}, "1", "made up"},
		{"ns.size.n", []string{"invented"}, "1", "made up"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Listing = %+v, want %+v", got, want)
	}
	got, err = d.Listing([]byte("ns.color\treal\tthe palette\nns.size\tinvented\tmade up\nns.size.n\tvalid\tcounted\n"), ListingState)
	if err != nil {
		t.Fatal(err)
	}
	if want := (Path{"ns.size.n", []string{"valid"}, "1", "counted"}); len(got) != 3 || !reflect.DeepEqual(got[2], want) {
		t.Errorf("Listing with a row for ns.size.n = %+v, want its last path %+v", got, want)
	}
}

func TestListingGivesAFieldReachedThroughALinkedTableTheFieldsOwnRow(t *testing.T) {
	d := loaded(t, module("ns", map[string]string{
		"r.json": `{"format":"{name}","rows":"r.tsv","key":"code","name":"name"}`,
		"r.tsv":  "code\tname\n1\ta\n",
		"m.json": `{"format":"{name}","rows":"m.tsv","key":"code","parent":"r"}`,
		"m.tsv":  "code\tname\tr\n10\tc\t1\n",
	}))
	got, err := d.Listing([]byte("ns.m\treal\tthe m register\nns.m.name\tvalid\tthe name office\nns.r\tinvented\tmade up\n"), ListingState)
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, p := range got {
		sources[p.Path] = strings.Join(p.Kind, ", ") + " from " + p.Source
	}
	for path, want := range map[string]string{
		"ns.m.name":   "valid from the name office",
		"ns.r.m.name": "valid from the name office",
		"ns.r.m.code": "real from the m register",
		"ns.r.m":      "real from the m register",
		"ns.r.name":   "invented from made up",
	} {
		if sources[path] != want {
			t.Errorf("%s is %q, want %q", path, sources[path], want)
		}
	}
}

func TestListingGivesALinkedTableItsOwnRowBesideAColumnOfItsName(t *testing.T) {
	d := loaded(t, module("ns", map[string]string{
		"r.json": `{"format":"{name}","rows":"r.tsv","key":"code","name":"name"}`,
		"r.tsv":  "code\tname\n1\ta\n",
		"m.json": `{"format":"{name}","rows":"m.tsv","key":"code","parent":"r"}`,
		"m.tsv":  "code\tname\tm\tr\n10\tc\tx\t1\n",
	}))
	got, err := d.Listing([]byte("ns.m\treal\tthe m register\nns.m.m\tvalid\tthe m office\nns.m.name\tinvented\tthe name office\nns.r\tinvented\tmade up\n"), ListingState)
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, p := range got {
		sources[p.Path] = p.Source
	}
	for path, want := range map[string]string{
		"ns.m":        "the m register",
		"ns.m.m":      "the m office",
		"ns.r.m":      "the m register",
		"ns.r.m.m":    "the m office",
		"ns.r.m.name": "the name office",
		"ns.r.m.code": "the m register",
	} {
		if sources[path] != want {
			t.Errorf("the source of %s is %q, want %q", path, sources[path], want)
		}
	}
}

func TestLoadRefusesAModuleMountedInsideAnother(t *testing.T) {
	inner, outer := module("ns.sub", map[string]string{"b.json": `"y"`}), module("ns", map[string]string{"a.json": `"x"`})
	for _, modules := range [][]fs.FS{{outer, inner}, {inner, outer}} {
		_, err := Load(modules)
		if err == nil || !strings.Contains(err.Error(), `"ns.sub"`) || !strings.Contains(err.Error(), `"ns"`) {
			t.Errorf("Load = %v, want an error naming the namespaces \"ns\" and \"ns.sub\"", err)
		}
	}
	if _, err := Load([]fs.FS{module("ns", map[string]string{"a.json": `"x"`}), module("nsb", map[string]string{"b.json": `"y"`})}); err != nil {
		t.Errorf("Load of ns and nsb = %v, want both loaded: neither lies inside the other", err)
	}
}

func TestListingNamesEveryMistakeInTheKinds(t *testing.T) {
	const file = "data-listing/kinds-and-sources.tsv"
	whole := "ns.color\treal\tthe palette\nns.size\tinvented\tmade up\n"
	for _, c := range []struct {
		name, kinds string
		want        []string
	}{
		{"a row listed twice", whole + "ns.color\treal\tthe palette\n", []string{file + ":3: ns.color is listed twice"}},
		{"kinds out of order", "ns.color\tvalid, real\tthe palette\nns.size\tinvented\tmade up\n", []string{file + ":1: kinds are real, valid, invented, each at most once and in that order"}},
		{"a kind twice", "ns.color\treal\tthe palette\nns.size\tinvented, invented\tmade up\n", []string{file + ":2: kinds are real, valid, invented, each at most once and in that order"}},
		{"an unknown kind", "ns.color\tfake\tthe palette\nns.size\tinvented\tmade up\n", []string{file + ":1: kinds are real, valid, invented, each at most once and in that order"}},
		{"a path with no row", "ns.color\treal\tthe palette\n", []string{"no row in " + file + " for ns.size", "no row in " + file + " for ns.size.n"}},
		{"a row with no shipped path", whole + "ns.shade\treal\tthe palette\n", []string{file + ": ns.shade is no shipped path"}},
		{"a row with no source", "ns.color\treal\nns.size\tinvented\tmade up\n", []string{file + ":1: want path, kinds and source, tab-separated, none empty", "no row in " + file + " for ns.color"}},
		{"two mistakes at once", "ns.color\tvalid, real\tthe palette\n", []string{file + ":1: kinds are", "no row in " + file + " for ns.size"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			listed, err := colorsAndSizes(t).Listing([]byte(c.kinds), ListingState)
			if err == nil {
				t.Fatalf("Listing = %+v, want an error", listed)
			}
			for _, want := range c.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Listing's error lacks %q:\n%v", want, err)
				}
			}
		})
	}
}

func TestExampleCell(t *testing.T) {
	for _, c := range []struct{ name, example, want string }{
		{"plain", "red", "`red`"},
		{"empty", "", "(empty)"},
		{"48 runes", strings.Repeat("é", 48), "`" + strings.Repeat("é", 48) + "`"},
		{"49 runes", strings.Repeat("é", 49), "`" + strings.Repeat("é", 48) + "…`"},
		{"a line break", "a\r\nb", "`a ⏎ b`"},
		{"a pipe", "a|b", "`a\\|b`"},
		{"a backtick", "a`b", "`` a`b ``"},
	} {
		if got := exampleCell(c.example); got != c.want {
			t.Errorf("exampleCell of %s = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestListingFilesHoldEveryPathUnderItsModule(t *testing.T) {
	d := loaded(t,
		module("zz", map[string]string{"b.json": `"y|z"`}),
		module("ns", map[string]string{"a.json": `"x"`}),
		module("empty", nil),
	)
	listed, err := d.Listing([]byte("ns.a\treal\ta | b\nzz.b\tinvented\tmade up\n"), ListingState)
	if err != nil {
		t.Fatal(err)
	}
	files := d.ListingFiles(listed)
	if len(files) != 2 || files[0].Name != PathsFile || files[1].Name != DocFile {
		t.Fatalf("ListingFiles = %+v, want %s then %s", files, PathsFile, DocFile)
	}
	wantJSON := "[\n\t{\"path\":\"ns.a\",\"kind\":[\"real\"],\"example\":\"x\",\"source\":\"a | b\"},\n\t{\"path\":\"zz.b\",\"kind\":[\"invented\"],\"example\":\"y|z\",\"source\":\"made up\"}\n]\n"
	if got := string(files[0].Body); got != wantJSON {
		t.Errorf("%s =\n%s\nwant\n%s", PathsFile, got, wantJSON)
	}
	wantTables := "<details>\n<summary><code>ns</code>, 1 paths</summary>\n\n| Path | Kind | Example | Source |\n|---|---|---|---|\n| `ns.a` | real | `x` | a \\| b |\n\n</details>\n\n" +
		"<details>\n<summary><code>zz</code>, 1 paths</summary>\n\n| Path | Kind | Example | Source |\n|---|---|---|---|\n| `zz.b` | invented | `y\\|z` | made up |\n\n</details>\n"
	if got := string(files[1].Body); !strings.HasSuffix(got, "\n\n"+wantTables) || strings.Contains(got, "empty") {
		t.Errorf("%s =\n%s\nwant it to end with a table per module that holds a path, sorted:\n%s", DocFile, got, wantTables)
	}
}

func TestManifestsAndReadsFollowWhatAModuleReads(t *testing.T) {
	d := loaded(t,
		module("one", map[string]string{"a.json": `"{/two.b}"`}),
		module("two", map[string]string{"b.json": `"{/sub.three.c}"`}),
		module("sub.three", map[string]string{"c.json": `"z"`}),
	)
	reads, err := d.Reads()
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string][]string{"one": {"sub.three", "two"}, "two": {"sub.three"}, "sub.three": nil}; !reflect.DeepEqual(reads, want) {
		t.Errorf("Reads = %v, want %v", reads, want)
	}
	manifests, err := d.Manifests()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, m := range manifests {
		names = append(names, m.Name)
	}
	if want := []string{"data/one/.fejkdata.json", "data/two/.fejkdata.json", "data/sub/three/.fejkdata.json"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("Manifests names %v, want %v", names, want)
	}
	if got := string(manifests[2].Body); !strings.HasPrefix(got, "{\n\t\"index\": {\n\t\t\"sub.three.c\": {") || !strings.HasSuffix(got, "\n\t},\n\t\"namespace\": \"sub.three\"\n}\n") {
		t.Errorf("the manifest of sub.three =\n%s\nwant an index of sub.three.c, then the namespace", got)
	}
}

func TestShapeNamesReads(t *testing.T) {
	d := loaded(t, module("ns", map[string]string{
		"a.json":     `{"format":"{x}","x":["{/ns.b}",{"format":"{/ns.c.v}","weight":2}]}`,
		"b.json":     `"y"`,
		"c.json":     `{"format":"{v} {n}","v":["z","w"],"n":[null,{"format":"{int(1,9)}","datatype":"integer"}]}`,
		"d/pos.json": `["{.q}","{/ns.b}"]`,
		"d/q.json":   `"r"`,
		"e.json":     `"{/ns.a}"`,
	}))
	want := "ns.a\tformat \"{x}\"\treads ns.b ns.c\tcolumns x\nns.a.x\tstring\nns.b\tformat \"y\"\nns.c\tformat \"{v} {n}\"\tcolumns v n\nns.c.n\tinteger null\nns.c.v\tstring\nns.d.pos\treads ns.b ns.d.q\nns.d.q\tformat \"r\"\nns.e\tformat \"{/ns.a}\"\treads ns.a\n"
	if got := d.Shape(); got.Name != ShapeFile || string(got.Body) != want {
		t.Fatalf("Shape = %s\n%s\nwant %s\n%s", got.Name, got.Body, ShapeFile, want)
	}
}

func TestShapeNamesTables(t *testing.T) {
	d := loaded(t, module("ns", map[string]string{
		"w.json": `"x"`,
		"r.json": `{"format":"{name}","rows":"r.tsv","key":"code","name":"name","weight":"w"}`,
		"r.tsv":  "code\tname\tw\n1\ta\t2\n2\tb\t3\n",
		"m.json": `{"format":"{name} {/ns.w}","rows":"m.tsv","key":"code","parent":"r"}`,
		"m.tsv":  "name\tcode\tr\nc\t10\t1\nd\t20\t2\n",
	}))
	want := "ns.m\tformat \"{name} {/ns.w}\"\tkey code\tparent r\treads ns.w\tcolumns name code r\nns.m.code\tstring\nns.m.name\tstring\nns.m.r\tstring\nns.r\tformat \"{name}\"\tkey code\tname name\tweight w\tcolumns code name w\nns.r.code\tstring\nns.r.m\nns.r.m.code\nns.r.m.name\nns.r.m.r\nns.r.name\tstring\nns.r.w\tstring\nns.w\tformat \"x\"\n"
	if got := string(d.Shape().Body); got != want {
		t.Fatalf("Shape =\n%s\nwant\n%s", got, want)
	}
}

func TestListingShapeNamesEachKeyAndKindWord(t *testing.T) {
	want := "key path\nkey kind\nkey example\nkey source\nkind real\nkind valid\nkind invented\n"
	if got := ListingShape(); got.Name != ListingShapeFile || string(got.Body) != want {
		t.Errorf("ListingShape = %s\n%s\nwant %s\n%s", got.Name, got.Body, ListingShapeFile, want)
	}
}
