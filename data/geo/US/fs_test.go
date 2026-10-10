package US_test

import (
	"io/fs"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/data/geo/US"
)

func TestFSLoadsByItsManifest(t *testing.T) {
	if _, err := fs.Stat(US.FS, ".fejkdata.json"); err != nil {
		t.Fatalf("FS holds no manifest: %v", err)
	}
	f, err := fejkdata.New(fejkdata.WithDataFS(US.FS))
	if err != nil {
		t.Fatalf("New(WithDataFS(FS)) = %v", err)
	}
	if len(f.List()) == 0 {
		t.Error("List() is empty, want the module's categories")
	}
}
