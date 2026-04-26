// Command agent-scan is the agent-scan CLI entrypoint. It dispatches the
// `quality`, `security`, `peer`, and `version` subcommands documented in
// docs/agent-scan-plan.md. Subcommand wiring beyond `version` is added in
// later steps; this scaffold proves the toolchain and exposes [Run] for
// testing.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/pwilliams-ck/agent-scan/internal/version"
)

// usage is written to stderr when invoked with no subcommand or an
// unknown one. It intentionally lists every subcommand the plan defines,
// even those not yet wired, so the surface stays visible while the rest
// of the build sequence fills it in.
const usage = `usage: agent-scan <command> [flags]

commands:
  quality   run the quality profile against the current scope
  security  run the security profile against the current scope
  peer      dispatch a peer audit to claude or codex
  version   print the agent-scan build version

See docs/agent-scan-plan.md for the full flag reference.
`

func main() {
	os.Exit(Run(os.Args, os.Stdout, os.Stderr))
}

// Run executes one CLI invocation and returns the process exit code.
//
// argv is the full process argv (argv[0] is the program name). stdout
// and stderr receive normal and error output respectively; callers pass
// their own [io.Writer]s rather than reaching for [os.Stdout] /
// [os.Stderr] so behavior is observable from tests.
//
// Exit codes follow docs/agent-scan-plan.md:
//
//	0  pass or warning
//	1  blocking findings or failed checks
//	2  scanner/runtime error such as bad args or no git repo
//	3  peer recursion blocked
func Run(argv []string, stdout, stderr io.Writer) int {
	if len(argv) < 2 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch argv[1] {
	case "version":
		fmt.Fprintf(stdout, "agent-scan %s\n", version.String())
		return 0
	default:
		fmt.Fprintf(stderr, "agent-scan: unknown command %q\n\n", argv[1])
		fmt.Fprint(stderr, usage)
		return 2
	}
}
