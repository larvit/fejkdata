package geono_test

import (
	"io/fs"
	"testing"

	"github.com/larvit/fejkdata"
	geono "github.com/larvit/fejkdata/data/geo/NO"
)

func TestFSLoadsByItsManifest(t *testing.T) {
	if _, err := fs.Stat(geono.FS, ".fejkdata.json"); err != nil {
		t.Fatalf("FS holds no manifest: %v", err)
	}
	f, err := fejkdata.New(fejkdata.WithDataFS(geono.Modules()...), fejkdata.WithSeed(1))
	if err != nil {
		t.Fatalf("New(WithDataFS(Modules()...)) = %v", err)
	}
	if len(f.List()) == 0 {
		t.Error("List() is empty, want the module's categories")
	}
}
