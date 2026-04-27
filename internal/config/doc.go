// Package config loads agent-scan's resolved configuration.
//
// Precedence, lowest to highest:
//
//  1. Built-in defaults from [Defaults].
//  2. Global config at <home>/.config/agent-scan/config.json.
//  3. Repo config at <repo>/.agent-scan.json.
//  4. CLI flags via [Flags].
//
// JSON is the only on-disk format. Decoding uses
// [encoding/json.Decoder.DisallowUnknownFields] so a typo in any layer
// fails loudly with the offending field name and file path rather than
// being silently ignored. Missing files at any layer are not errors —
// the layer simply contributes nothing.
//
// Home and repo paths are arguments to [Load]; the package never
// consults the process environment or [os.UserHomeDir] so tests can
// build hermetic fake filesystems with [testing.T.TempDir].
package config
