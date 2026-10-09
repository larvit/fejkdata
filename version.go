package fejkdata

//go:generate env REGENERATE=1 go test -count=1 -run ^TestVersionIsTheNewestChangelogHeading$ .

// Version is the newest release CHANGELOG.md heads, or v0.0.0 before the first; a build between releases builds on it.
const Version = "v0.0.0"
