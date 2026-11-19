package fejkdata

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var (
	archiveFiles  = regexp.MustCompile(`(?m)^      - (?:src: )?(\S+)$`)
	packagingPath = regexp.MustCompile(`packaging=\$\(changed ([^)]*)\)`)
)

func TestPackagingFilterCoversTheFilesAnArchiveShips(t *testing.T) {
	goreleaser, err := os.ReadFile("release-tooling/goreleaser.yaml")
	if err != nil {
		t.Fatal(err)
	}
	archives, _, _ := strings.Cut(string(goreleaser), "\nchecksum:")
	_, archives, _ = strings.Cut(archives, "    files:\n")
	workflow, err := os.ReadFile(".github/workflows/test.yml")
	if err != nil {
		t.Fatal(err)
	}
	filter := packagingPath.FindStringSubmatch(string(workflow))
	if filter == nil {
		t.Fatal("test.yml sets no packaging= filter")
	}
	paths := strings.Fields(strings.ReplaceAll(filter[1], "'", ""))
	shipped := archiveFiles.FindAllStringSubmatch(archives, -1)
	if len(shipped) < 4 {
		t.Fatalf("found only %d archive files in goreleaser.yaml", len(shipped))
	}
	for _, m := range shipped {
		if !coveredBy(m[1], paths) {
			t.Errorf("an archive ships %s, which test.yml's packaging filter %v does not cover, so a change to it skips the Packages dry run", m[1], paths)
		}
	}
}

func coveredBy(file string, paths []string) bool {
	for _, p := range paths {
		if file == p || strings.HasPrefix(file, p+"/") {
			return true
		}
		if strings.HasPrefix(p, "*.") && strings.HasSuffix(file, p[1:]) {
			return true
		}
	}
	return false
}
