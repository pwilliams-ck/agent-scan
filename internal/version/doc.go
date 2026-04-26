// Package version reports the agent-scan build version.
//
// The string returned by [String] defaults to "dev" for local builds and is
// overridden at release time via the linker flag
// -ldflags '-X github.com/pwilliams-ck/agent-scan/internal/version.value=vX.Y.Z'.
package version
