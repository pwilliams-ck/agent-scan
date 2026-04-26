# 08 — Peer dispatcher

> Goal: `internal/peer/` constructs the right child invocation for Claude
> and Codex, enforces the recursion guard, and is integration-tested with
> fake binaries on PATH (mocks alone are not sufficient).

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] `BuildArgs` tests for `codex` and `claude` written and red; argv
      slices match the verified flag syntax from step 01
- [ ] Recursion-guard test written and red
- [ ] Integration tests with fake `claude` and `codex` binaries on PATH
      written and red (these run real subprocesses, not mocks)
- [ ] All peer tests green, including `-race`
- [ ] `internal/peer/` exports `Dispatch`, `EnvDepth`, `ErrPeerRecursion`;
      mapping `ErrPeerRecursion → exit 3` is wired in `cmd/agent-scan/main.go`
- [ ] Prompt constants live in `prompts.go` with doc comments naming each
      forbidden action (peer call, edit, git mutation) and why each forbid
      exists
- [ ] Manual smoke: `AGENT_SCAN_PEER_DEPTH=1 agent-scan peer codex security`
      exits 3 with a clear stderr message
- [ ] `go test ./internal/peer/... -count=1 -race`, `go vet`,
      `gofmt -l internal/peer/` all clean

Commit each box as it flips.

## Tests to write first (red)

| Test | Asserts |
|------|---------|
| `TestBuildArgs_Codex_MatchesExpected` | argv equals expected slice (asserts step-01 flags) |
| `TestBuildArgs_Claude_MatchesExpected` | argv equals expected slice |
| `TestDispatch_DepthAlreadySet_ReturnsErrPeerRecursion` | `t.Setenv(EnvDepth, "1")` → `errors.Is(err, ErrPeerRecursion)` |
| `TestDispatch_ChildEnvIncludesDepthOne` | fake binary writes its env to a file; assert `AGENT_SCAN_PEER_DEPTH=1` present |
| `TestDispatch_ChildEnvDoesNotLeakAgentScanFlags` | child env scrubbed |
| `TestDispatch_TimeoutEnforced` | fake binary sleeps; ctx with 100ms deadline → `context.DeadlineExceeded` |
| `TestDispatch_NonZeroExit_PropagatesAsError` | child exit code surfaces in error |
| `TestDispatch_Codex_Integration_FakeBinary` | shell-script `codex` in `t.TempDir`; PATH prepended; stdout flows through |
| `TestDispatch_Claude_Integration_FakeBinary` | shell-script `claude`; same |

The integration tests are the load-bearing piece. Mocks would let argv
typos slip through unnoticed.

## Files touched

- `internal/peer/{peer.go, peer_test.go, prompts.go, doc.go}`
- `cmd/agent-scan/main.go` — wire the `peer` subcommand and map
  `ErrPeerRecursion` to exit 3

## Documentation requirements

- Package comment explains the recursion guard rationale and the
  read-only-by-design contract (peer is one level deep, period).
- `Dispatch` doc comment lists every error it can return and the exit code
  each maps to.
- Prompt constants have a doc comment explaining what's intentionally
  forbidden (peer call, edits, git mutations) and why removing any one of
  those forbids would break the contract.
- One inline comment above `EnvDepth` explaining the value `"1"` is not
  load-bearing — *presence* is the signal.

## Acceptance check

```bash
go test ./internal/peer/... -count=1 -race
go test ./internal/peer/... -run Integration -count=1
```

End-to-end (real binaries installed, after step 10): in a fixture repo,
`agent-scan peer codex security` returns a Codex audit report;
`AGENT_SCAN_PEER_DEPTH=1 agent-scan peer codex security` exits 3.
