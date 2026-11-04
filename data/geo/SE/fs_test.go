package geose_test

import (
	"io/fs"
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	geose "github.com/larvit/fejkdata/data/geo/SE"
)

func TestFSLoadsByItsManifest(t *testing.T) {
	if _, err := fs.Stat(geose.FS, ".fejkdata.json"); err != nil {
		t.Fatalf("FS holds no manifest: %v", err)
	}
	f, err := fejkdata.New(fejkdata.WithDataFS(geose.Modules()...), fejkdata.WithSeed(1), fejkdata.WithToday(time.Now()))
	if err != nil {
		t.Fatalf("New(WithDataFS(Modules()...)) = %v", err)
	}
	if len(f.List()) == 0 {
		t.Error("List() is empty, want the module's categories")
	}
}
