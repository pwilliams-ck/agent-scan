# 06 — Quality profile

> Goal: `internal/profiles/quality.go` runs generic + per-stack checks and
> emits the `Report` shape from step 05. Stack detection lives in its own
> package; subprocess execution lives in `internal/run/`.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] `internal/stacks/` tests written and red; then green
- [ ] `internal/run/Exec` extracted with its own tests (env scrubbing,
      timeout enforcement, exit-code propagation)
- [ ] All `internal/profiles/quality.go` tests written and red
- [ ] All quality tests green, including `-race`
- [ ] Missing tools → `SkippedTool`, not error (when
      `policy.missing_tools = "skip"`); with `"fail"`, → finding with
      `command_failed` severity
- [ ] Doc comments on every exported identifier; each `Check` documents the
      tool it requires, the severity it emits on failure, and what it
      asserts
- [ ] Manual smoke: a fixture repo with broken `gofmt` makes
      `agent-scan quality` exit 1 with a finding pointing at file:line
- [ ] `go test ./internal/{stacks,run,profiles}/... -count=1 -race`,
      `go vet`, `gofmt -l` all clean

Commit each box as it flips.

## Tests to write first (red)

`internal/stacks/stacks_test.go`:

| Test | Asserts |
|------|---------|
| `TestDetect_GoMarker` / `_NodeMarker` / `_PythonMarker` / `_RustMarker` / `_DockerMarker` | each marker file produces the right `Stack` |
| `TestDetect_MultipleStacks_AllDetected` | polyglot repo returns all matches |
| `TestDetect_NoMarkers_ReturnsEmpty` | clean repo |

`internal/run/run_test.go`:

| Test | Asserts |
|------|---------|
| `TestExec_RunsCommandAndCapturesStdout` | basic happy path |
| `TestExec_TimeoutCancels` | ctx deadline → `context.DeadlineExceeded` |
| `TestExec_NonZeroExit_ReturnsExitError` | exit code preserved |
| `TestExec_EnvIsExplicit_NotInherited` | only env passed in is visible to child |

`internal/profiles/quality_test.go`:

| Test | Asserts |
|------|---------|
| `TestRun_GenericChecks_AlwaysExecute` | empty repo (no stacks) still runs generic checks |
| `TestRun_GoStack_RunsGofmtAndVetWhenLintMissing` | fake `golangci-lint` absent on PATH |
| `TestRun_GoStack_RunsGolangciLintWhenPresent` | fake binary on PATH wins |
| `TestRun_MissingTool_ProducesSkippedNotError` | `policy.missing_tools = "skip"` |
| `TestRun_MissingTool_FailsWhenPolicyFail` | `"fail"` → finding |
| `TestRun_FailedCommand_ProducesBlockingFinding` | flips result to `fail` |
| `TestRun_RespectsTimeout` | long-running fake binary cancelled |
| `TestRun_IgnorePathsExcludesFiles` | `cfg.Ignore.Paths` honoured |

Subprocess tests use `t.Setenv("PATH", tmpBin)` plus shell scripts
written into `tmpBin` to fake external tools per test.

## Files touched

- `internal/stacks/{stacks.go, stacks_test.go, doc.go}`
- `internal/run/{run.go, run_test.go, doc.go}`
- `internal/profiles/{quality.go, quality_test.go, doc.go}`
- `internal/profiles/checks/*.go` — one file per check family
  (e.g. `gofmt.go`, `gotest.go`, `npm.go`, `cargo.go`)

## Documentation requirements

- `Stack` enum values commented with their detection marker.
- Each check documents: tool/binary required, severity on failure, what
  it asserts.
- `Run` documents that individual check failures become findings; only
  environmental failures (e.g. cannot resolve repo root) return error.
- `internal/run/Exec` documents the env-explicit contract loudly: the
  child process inherits *only* the env passed in. No `os.Environ()`
  leakage.

## Acceptance check

```bash
go test ./internal/stacks/... ./internal/run/... ./internal/profiles/... \
    -count=1 -race
```

Manual smoke (after the binary builds): in a fixture repo with broken
`gofmt`, `agent-scan quality` exits 1 with a finding pointing at
`file:line`.
