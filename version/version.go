package version

// Set with LDFLAGS.
var (
	// Version of the release, the value injected by .goreleaser
	version = `{{.Version}}`

	// Commit hash of the release, the value injected by .goreleaser
	commit = `{{.Commit}}`
)
func GetVersion() string {
	return version
}

// GetFormattedVersion returns the current version and commit hash
func GetFormattedVersion() string {
	return fmt.Sprintf("%s (%s)", version, commit)
}
