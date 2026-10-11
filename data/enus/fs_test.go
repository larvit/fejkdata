package enus_test

import (
	"io/fs"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/data/enus"
)

func TestFSLoadsByItsManifest(t *testing.T) {
	if _, err := fs.Stat(enus.FS, ".fejkdata.json"); err != nil {
		t.Fatalf("FS holds no manifest: %v", err)
	}
	f, err := fejkdata.New(fejkdata.WithDataFS(enus.FS))
	if err != nil {
		t.Fatalf("New(WithDataFS(FS)) = %v", err)
	}
	if len(f.List()) == 0 {
		t.Error("List() is empty, want the module's categories")
	}
}
