# 02 — Scaffold the Go module

> Goal: minimum compilable `agent-scan` binary that prints a version string.
> No business logic. Proves the toolchain, layout, and TDD harness are wired
> up.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] Tests for `version.String` and `Run` written; all red for the right
      reason (failing assertion, not a compile error)
- [ ] All tests green
- [ ] `cmd/agent-scan/main.go` and `internal/version/version.go` exist;
      `go build ./cmd/agent-scan` produces a working binary
- [ ] Doc comments on every exported identifier (`revive` `exported` clean)
- [ ] Package comments on `main` and `internal/version`
- [ ] `go test ./...`, `go vet ./...`, `gofmt -l .` all clean
- [ ] `./agent-scan version` prints `agent-scan dev` (or the build-flag
      value); `./agent-scan` with no args exits 2 with usage on stderr

Commit each box as it flips.

## Tests to write first (red)

| Test | Asserts |
|------|---------|
| `internal/version: TestString_NotEmpty` | `version.String()` returns non-empty, starts with `v` or equals `dev` |
| `cmd/agent-scan: TestRun_Version_PrintsAgentScanLine` | `Run(["agent-scan","version"])` writes `agent-scan ` to stdout, returns 0 |
| `cmd/agent-scan: TestRun_NoArgs_PrintsUsageAndExitsTwo` | usage on stderr, returns 2 |
| `cmd/agent-scan: TestRun_UnknownSubcommand_PrintsUsageAndExitsTwo` | usage on stderr, returns 2 |

Drive `main` with `os.Exit(Run(os.Args))` so `Run` is testable without
forking. `Run` takes its own `io.Writer` for stdout/stderr — don't reach for
the package-level `os.Stdout`.

## Files touched

- `go.mod` (new)
- `cmd/agent-scan/{main.go, main_test.go}`
- `internal/version/{version.go, version_test.go, doc.go}`

## Documentation requirements

- Package comment on `main` explains this is the CLI entrypoint and points
  readers at `docs/agent-scan-plan.md`.
- Package comment on `internal/version` names the convention: build-flag
  populated via `-ldflags '-X .../version.value=v0.0.1'`, defaults to
  `"dev"`.
- `version.String` doc comment notes how the value is set.

## Acceptance check

```bash
go build ./...
go test ./...
go vet ./...
gofmt -l .                                # empty output
go build -o ./agent-scan ./cmd/agent-scan
./agent-scan version                      # prints "agent-scan dev"
./agent-scan; echo $?                     # usage on stderr, prints 2
```
