package datafiles

import (
	"errors"
	"os"
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
		got = append(got, handed{c.Folders, c.Name, c.JSON.Any()})
		return nil
	})
	return got, err
}

func TestWalkHandsEachCategoryInOrder(t *testing.T) {
	got, err := walk(t, Source{fsys: fstest.MapFS{
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
	got, err := walk(t, Source{fsys: fstest.MapFS{"data/a.json": {Data: []byte(`"x"`)}, "b.json": {Data: []byte(`"y"`)}}, base: "data"})
	if err != nil || len(got) != 1 || got[0].name != "a" || got[0].dir != nil {
		t.Errorf("Walk = %#v, %v, want a alone at the base", got, err)
	}
}

func TestReadRowsReadsOnlyAFileBesideTheCategory(t *testing.T) {
	src := Source{fsys: fstest.MapFS{
		"t.json":     {Data: []byte(`{"rows":"t.tsv"}`)},
		"t.tsv":      {Data: []byte("key\nA\n")},
		"sub/u.tsv":  {Data: []byte("key\nB\n")},
		"sub/u.json": {Data: []byte(`{"rows":"u.tsv"}`)},
	}, label: "lbl"}
	rows := map[string]string{}
	err := src.Walk(func(c Category) error {
		data, err := c.ReadRows(c.Name + ".tsv")
		rows[c.Name] = data
		if err != nil {
			return err
		}
		if _, err := c.ReadRows("x.tsv"); err == nil || err.Error() != "rows names x.tsv, which is not beside it in "+map[string]string{"t": "lbl", "u": "lbl/sub"}[c.Name] {
			t.Errorf("ReadRows(x.tsv) = %v, want an error saying x.tsv is not beside %s", err, c.Name)
		}
		return nil
	})
	if err != nil || rows["t"] != "key\nA\n" || rows["u"] != "key\nB\n" {
		t.Errorf("Walk = %v, read %q", err, rows)
	}
}

func TestWalkIgnoresARowsFileNoCategoryNames(t *testing.T) {
	if _, err := walk(t, Source{fsys: fstest.MapFS{"a.json": {Data: []byte(`"x"`)}, "a.tsv": {Data: []byte("key\n")}}, label: "lbl"}); err != nil {
		t.Errorf("Walk = %v, want a.tsv ignored", err)
	}
}

func TestWalkLabelsWhatACategoryFileGetsWrong(t *testing.T) {
	for _, c := range []struct {
		name    string
		fs      fstest.MapFS
		compile error
		want    string
	}{
		{"its compile", fstest.MapFS{"a.json": {Data: []byte(`"x"`)}}, errors.New("bad"), "lbl/a.json: bad"},
		{"its JSON", fstest.MapFS{"a.json": {Data: []byte(`{`)}}, nil, "lbl/a.json: unexpected end of JSON input"},
		{"its name", fstest.MapFS{"a.b.json": {Data: []byte(`"x"`)}}, nil, "lbl/a.b.json: category "},
		{"its folder's name", fstest.MapFS{"a.b/c.json": {Data: []byte(`"x"`)}}, nil, "lbl/a.b: folder "},
	} {
		err := Source{fsys: c.fs, label: "lbl"}.Walk(func(Category) error { return c.compile })
		if err == nil || !strings.HasPrefix(err.Error(), c.want) {
			t.Errorf("%s: Walk = %v, want it to start %q", c.name, err, c.want)
		}
	}
}

func TestWalkChecksAFolderNameOnlyWhenItHoldsData(t *testing.T) {
	if _, err := walk(t, Source{fsys: fstest.MapFS{"a.b/notes.txt": {Data: []byte("x")}, "c.json": {Data: []byte(`"x"`)}}}); err != nil {
		t.Errorf("Walk = %v, want a folder holding no data skipped whatever its name", err)
	}
}

func TestWalkRefusesADiskPathThatIsNoDirectory(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct{ path, want string }{
		{"", "a data path is empty"},
		{filepath.Join(dir, "none"), "data path " + filepath.Join(dir, "none") + ": "},
	} {
		_, err := walk(t, Dir(c.path))
		if err == nil || !strings.HasPrefix(err.Error(), c.want) {
			t.Errorf("Walk(%q) = %v, want it to start %q", c.path, err, c.want)
		}
	}
}

func TestLoadHandsOneCategoryAndItsRows(t *testing.T) {
	src := Source{fsys: fstest.MapFS{
		"data/sub/t.json": {Data: []byte(`{"rows":"t.tsv"}`)},
		"data/sub/t.tsv":  {Data: []byte("key\nA\n")},
		"data/sub/u.json": {Data: []byte(`{`)},
	}, base: "data"}
	var got handed
	var rows string
	err := src.Load([]string{"sub"}, "t", func(c Category) error {
		got = handed{c.Folders, c.Name, c.JSON.Any()}
		var err error
		rows, err = c.ReadRows("t.tsv")
		return err
	})
	if err != nil || got.name != "t" || !reflect.DeepEqual(got.dir, []string{"sub"}) || rows != "key\nA\n" {
		t.Errorf("Load = %v, handed %#v with rows %q", err, got, rows)
	}
}

func TestDirReadsTheDirectoryAndLabelsItsErrorsWithIt(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.json"), []byte(`{`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := walk(t, Dir(dir))
	if err == nil || !strings.HasPrefix(err.Error(), filepath.ToSlash(filepath.Join(dir, "a.json"))+": ") {
		t.Errorf("Walk = %v, want the error labelled with the directory", err)
	}
	if _, err := walk(t, Dir(filepath.Join(dir, "a.json"))); err == nil || err.Error() != filepath.Join(dir, "a.json")+" is not a directory" {
		t.Errorf("Walk = %v, want a file refused as no directory", err)
	}
}
