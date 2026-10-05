package datafiles

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

type handed struct {
	dir  []string
	name string
	json any
}

func walk(t *testing.T, src Source) ([]handed, error) {
	t.Helper()
	var got []handed
	err := src.Walk(func(c Category) error {
		got = append(got, handed{c.Dir, c.Name, c.JSON})
		return nil
	})
	return got, err
}

func TestWalkHandsEachCategoryInOrder(t *testing.T) {
	got, err := walk(t, Source{FS: fstest.MapFS{
		"a.json":          {Data: []byte(`{"x":"1"}`)},
		"sub/b.json":      {Data: []byte(`"y"`)},
		"sub/notes.txt":   {Data: []byte("not data")},
		"empty/x.txt":     {Data: []byte("not data")},
		".git/c.json":     {Data: []byte(`"hidden"`)},
		"sub/.d.json":     {Data: []byte(`"hidden"`)},
		"sub/deep/e.json": {Data: []byte(`"z"`)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []handed{
		{nil, "a", map[string]any{"x": "1"}},
		{[]string{"sub"}, "b", "y"},
		{[]string{"sub", "deep"}, "e", "z"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Walk handed %#v, want %#v", got, want)
	}
}

func TestWalkStartsAtTheBaseDir(t *testing.T) {
	got, err := walk(t, Source{FS: fstest.MapFS{"data/a.json": {Data: []byte(`"x"`)}, "b.json": {Data: []byte(`"y"`)}}, BaseDir: "data"})
	if err != nil || len(got) != 1 || got[0].name != "a" || got[0].dir != nil {
		t.Errorf("Walk = %#v, %v, want a alone at the base", got, err)
	}
}

func TestReadRowsReadsOnlyAFileBesideTheCategory(t *testing.T) {
	src := Source{FS: fstest.MapFS{
		"t.json":     {Data: []byte(`{"rows":"t.tsv"}`)},
		"t.tsv":      {Data: []byte("key\nA\n")},
		"sub/u.tsv":  {Data: []byte("key\nB\n")},
		"sub/u.json": {Data: []byte(`{"rows":"u.tsv"}`)},
	}, Label: "lbl"}
	rows := map[string]string{}
	err := src.Walk(func(c Category) error {
		data, err := c.ReadRows(c.Name + ".tsv")
		rows[c.Name] = data
		if err != nil {
			return err
		}
		if _, err := c.ReadRows("x.tsv"); err == nil || err.Error() != "rows names x.tsv, which is not beside it in "+map[string]string{"t": "lbl", "u": "lbl/sub"}[c.Name] {
			t.Errorf("ReadRows(x.tsv) = %v, want the file named missing beside %s", err, c.Name)
		}
		return nil
	})
	if err != nil || rows["t"] != "key\nA\n" || rows["u"] != "key\nB\n" {
		t.Errorf("Walk = %v, read %q", err, rows)
	}
}

func TestWalkRefusesARowsFileNoCategoryNames(t *testing.T) {
	_, err := walk(t, Source{FS: fstest.MapFS{"a.json": {Data: []byte(`"x"`)}, "a.tsv": {Data: []byte("key\n")}}, Label: "lbl"})
	if err == nil || !strings.HasPrefix(err.Error(), "lbl/a.tsv: no category names it in its rows") {
		t.Errorf("Walk = %v, want a.tsv refused as named by no category", err)
	}
}

func TestWalkLabelsWhatACategoryFileGetsWrong(t *testing.T) {
	for _, c := range []struct {
		name string
		fs   fstest.MapFS
		want string
	}{
		{"its compile", fstest.MapFS{"a.json": {Data: []byte(`"x"`)}}, "lbl/a.json: bad"},
		{"its JSON", fstest.MapFS{"a.json": {Data: []byte(`{`)}}, "lbl/a.json: unexpected end of JSON input"},
		{"its name", fstest.MapFS{"a b.json": {Data: []byte(`"x"`)}}, "lbl/a b.json: category "},
		{"its folder's name", fstest.MapFS{"a b/c.json": {Data: []byte(`"x"`)}}, "lbl/a b: folder "},
	} {
		err := Source{FS: c.fs, Label: "lbl"}.Walk(func(Category) error { return errors.New("bad") })
		if err == nil || !strings.HasPrefix(err.Error(), c.want) {
			t.Errorf("%s: Walk = %v, want it to start %q", c.name, err, c.want)
		}
	}
}

func TestWalkChecksAFolderNameOnlyWhenItHoldsData(t *testing.T) {
	if _, err := walk(t, Source{FS: fstest.MapFS{"a b/notes.txt": {Data: []byte("x")}, "c.json": {Data: []byte(`"x"`)}}}); err != nil {
		t.Errorf("Walk = %v, want a folder holding no data skipped whatever its name", err)
	}
}

func TestWalkRefusesADiskPathThatIsNoDirectory(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct{ path, want string }{
		{"", "a data path is empty"},
		{filepath.Join(dir, "none"), "data path " + filepath.Join(dir, "none") + ": "},
	} {
		_, err := walk(t, Source{FS: fstest.MapFS{}, Label: c.path, OnDisk: true})
		if err == nil || !strings.HasPrefix(err.Error(), c.want) {
			t.Errorf("Walk(%q) = %v, want it to start %q", c.path, err, c.want)
		}
	}
}

func TestLoadHandsOneCategoryAndItsRows(t *testing.T) {
	src := Source{FS: fstest.MapFS{
		"data/sub/t.json": {Data: []byte(`{"rows":"t.tsv"}`)},
		"data/sub/t.tsv":  {Data: []byte("key\nA\n")},
		"data/sub/u.json": {Data: []byte(`{`)},
	}, BaseDir: "data"}
	var got handed
	var rows string
	err := src.Load([]string{"sub"}, "t", func(c Category) error {
		got = handed{c.Dir, c.Name, c.JSON}
		var err error
		rows, err = c.ReadRows("t.tsv")
		return err
	})
	if err != nil || got.name != "t" || !reflect.DeepEqual(got.dir, []string{"sub"}) || rows != "key\nA\n" {
		t.Errorf("Load = %v, handed %#v with rows %q", err, got, rows)
	}
}
