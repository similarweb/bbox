package version

// Set with LDFLAGS.
var version = "unset-dev"

func GetVersion() string {
	return version
}
