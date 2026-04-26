# 07 — Security profile

> Goal: built-in fake-secret detection plus optional ecosystem audits, all
> funnelled through the same `Finding` shape from step 05.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] `internal/secrets/` tests written and red, including a false-positive
      guard test
- [ ] `internal/secrets/` tests green, including `-race`
- [ ] `internal/profiles/security.go` tests written and red
- [ ] Security profile tests green, including `-race`
- [ ] Each pattern in `secrets.go` has a code comment naming the credential
      type, the canonical regex source (e.g. AWS docs URL), and the
      severity it emits
- [ ] Doc comments on every exported identifier; package comment names
      the false-positive philosophy
- [ ] Manual smoke: dropping a fake AWS key into a fixture file makes
      `agent-scan security` flag a finding (medium or higher) with file:line
- [ ] `go test ./internal/secrets/... ./internal/profiles/... -count=1 -race`,
      `go vet`, `gofmt -l` all clean

Commit each box as it flips.

## Tests to write first (red)

`internal/secrets/secrets_test.go`:

| Test | Asserts |
|------|---------|
| `TestScan_AWSAccessKey_Detected` | `AKIA…` pattern caught |
| `TestScan_SlackToken_Detected` | `xox[bpoa]-…` caught |
| `TestScan_GitHubToken_Detected` | `ghp_…` / `gho_…` caught |
| `TestScan_PrivateKeyHeader_Detected` | `-----BEGIN … PRIVATE KEY-----` |
| `TestScan_HighEntropyBase64_Detected` | long base64-ish strings |
| `TestScan_LoremIpsum_NotDetected` | false-positive guard |
| `TestScan_BinaryFile_Skipped` | binary detection short-circuits |
| `TestScan_LineNumbersAreOneIndexed` | first line is line 1, not 0 |
| `TestScan_SeveritiesMapPerPattern` | severity table honoured |

`internal/profiles/security_test.go`:

| Test | Asserts |
|------|---------|
| `TestRun_BuiltinSecrets_AlwaysRun` | runs even with no ecosystem tools available |
| `TestRun_GitleaksMissing_Skipped` | absence → `SkippedTool` |
| `TestRun_GitleaksPresent_FindingsNormalised` | fake binary on PATH emits a fixture; output translated into `Finding` shape |
| `TestRun_DisabledTool_SkippedWithReason` | `security.disabled_tools` honoured even when binary present |
| `TestRun_RespectsIgnorePaths` | `cfg.Ignore.Paths` honoured for both built-ins and ecosystem tools |

Reuse `internal/run/Exec` from step 06; do not call `os/exec` directly.

## Files touched

- `internal/secrets/{secrets.go, secrets_test.go, doc.go}`
- `internal/profiles/{security.go, security_test.go}`
- `internal/profiles/parsers/*.go` — only if multiple ecosystem tools need
  output translation (extracted during refactor)

## Documentation requirements

- Package comment in `internal/secrets` names the false-positive philosophy:
  *"we err on the side of false positives but built-in patterns never emit
  below `medium` severity. Ecosystem tools own their own thresholds."*
- Each pattern's code comment includes: credential type, source URL, severity.
- One inline comment explaining why binary files are skipped (false positives
  + cost).
- `Run` documents normalisation: every finding from any tool is reshaped
  into the v1 `Finding` struct before emission.

## Acceptance check

```bash
go test ./internal/secrets/... ./internal/profiles/... -count=1 -race
```

Manual smoke: drop a fake `AKIAIOSFODNN7EXAMPLE` into a fixture file, run
`agent-scan security`, observe a finding ≥ medium with file:line.
