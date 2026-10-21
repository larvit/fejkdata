package nbno_test

import (
	"io/fs"
	"testing"

	"github.com/larvit/fejkdata"
	nbno "github.com/larvit/fejkdata/data/nb_NO"
)

func TestFSLoadsByItsManifest(t *testing.T) {
	if _, err := fs.Stat(nbno.FS, ".fejkdata.json"); err != nil {
		t.Fatalf("FS holds no manifest: %v", err)
	}
	f, err := fejkdata.New(fejkdata.WithDataFS(nbno.Modules()...), fejkdata.WithSeed(1))
	if err != nil {
		t.Fatalf("New(WithDataFS(Modules()...)) = %v", err)
	}
	if len(f.List()) == 0 {
		t.Error("List() is empty, want the module's categories")
	}
}
