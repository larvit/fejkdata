package fejkdata

import (
	"embed"

	"github.com/larvit/fejkdata/internal/datafiles"
)

//go:generate env REGENERATE=1 go test -run ^TestShippedManifestIsCurrent$ .

//go:embed data data/.fejkdata.json
var shippedFS embed.FS

var shippedSource = datafiles.FS(shippedFS, "data")
