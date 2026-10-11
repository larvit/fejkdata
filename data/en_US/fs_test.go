package enus_test

import (
	"io/fs"
	"testing"

	"github.com/larvit/fejkdata"
	enus "github.com/larvit/fejkdata/data/en_US"
)

func TestFSLoadsByItsManifest(t *testing.T) {
	if _, err := fs.Stat(enus.FS, ".fejkdata.json"); err != nil {
		t.Fatalf("FS holds no manifest: %v", err)
	}
	f, err := fejkdata.New(fejkdata.WithDataFS(enus.Modules()...))
	if err != nil {
		t.Fatalf("New(WithDataFS(Modules()...)) = %v", err)
	}
	if len(f.List()) == 0 {
		t.Error("List() is empty, want the module's categories")
	}
}
