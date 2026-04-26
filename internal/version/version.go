package version

// String returns the agent-scan build version.
//
// It returns "dev" for local builds. Release builds inject a value of the
// form "vX.Y.Z" via -ldflags '-X .../internal/version.value=vX.Y.Z'.
func String() string {
	return value
}

// value is the version string set at build time via -ldflags. It defaults
// to "dev" so locally-built binaries always report a non-empty version.
var value = "dev"
