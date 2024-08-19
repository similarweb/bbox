package version

import "fmt"

// Set with LDFLAGS.
var (
	// Version of the release, the value will be injected by .goreleaser
	version = ``

	// Commit hash of the release, the value  will be injected by .goreleaser
	commit = ` development`

	// Commit date of the release, the value will be injected by .goreleaser
	date = ``

	// BuiltBy of the release, the value will be injected by .goreleaser
	builtBy = `hasn't been built yet`
)

func GetVersion() string {
	return version
}

// GetFormattedVersion returns the current version and commit hash
func GetFormattedVersion() string {
	return fmt.Sprintf("%s (%s %s)\nbuilt by: %s", version, commit, date, builtBy)
}
