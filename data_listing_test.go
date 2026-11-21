package fejkdata

import (
	"os"
	"testing"

	"github.com/larvit/fejkdata/internal/shipped"
)

const githubRenderLimit = 500 * 1024

func TestDataListingIsCurrent(t *testing.T) {
	kinds, err := os.ReadFile(shipped.KindsFile)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := shippedData(t).Listing(kinds, shipped.ListingState)
	if err != nil {
		t.Fatalf("the listing is not whole:\n%v", err)
	}
	for _, f := range shippedData(t).ListingFiles(listed) {
		if len(f.Body) > githubRenderLimit {
			t.Errorf("%s is %d bytes, past the %d GitHub renders; split it per module", f.Name, len(f.Body), githubRenderLimit)
		}
		generatedFileIsCurrent(t, f, regenerate)
	}
}
