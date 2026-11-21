package fejkdata

import (
	"testing"

	"github.com/larvit/fejkdata/internal/shipped"
)

const repin = `repin: docker compose run --rm --user "$(id -u):$(id -g)" repin`

func TestShippedShapeIsPinned(t *testing.T) {
	generatedFileIsCurrent(t, shippedData(t).Shape(), "a removed, renamed or retyped line is a breaking change; add the CHANGELOG.md entry, then "+repin)
}

func TestTheListingsShapeIsPinned(t *testing.T) {
	generatedFileIsCurrent(t, shipped.ListingShape(), "a removed or renamed line is a major, as the README's Versioning says; add the CHANGELOG.md entry, then "+repin)
}
