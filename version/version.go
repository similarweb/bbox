package version

// Set with LDFLAGS
var version = "unset"

func Version() string {
	return version
}
