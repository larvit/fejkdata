package fejkdata

//go:generate env REGENERATE=1 go test -count=1 -run ^TestVersionIsTheNewestChangelogHeading$ .

// Version is the newest release CHANGELOG.md names, or v0.0.0 before the first; source committed after a release still reports that release.
const Version = "v0.0.0"
