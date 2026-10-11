package geous_test

import (
	"io/fs"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/data/geous"
)

func TestFSLoadsByItsManifest(t *testing.T) {
	if _, err := fs.Stat(geous.FS, ".fejkdata.json"); err != nil {
		t.Fatalf("FS holds no manifest: %v", err)
	}
	f, err := fejkdata.New(fejkdata.WithDataFS(geous.FS))
	if err != nil {
		t.Fatalf("New(WithDataFS(FS)) = %v", err)
	}
	if len(f.List()) == 0 {
		t.Error("List() is empty, want the module's categories")
	}
}
