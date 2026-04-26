# 04 — Config loading

> Goal: precedence-resolved config from defaults → global → repo → flags,
> using `encoding/json` with `DisallowUnknownFields()`.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] All config tests written; each fails for the right reason
- [ ] All config tests green, including `-race`
- [ ] `internal/config/` exports `Config`, `Load`, `Validate`, plus sentinel
      errors (`ErrUnknownField`, `ErrInvalidScope`, etc.)
- [ ] `json.Decoder.DisallowUnknownFields()` enforced; typos surface with
      file path and field name
- [ ] Layered merge implemented and unit-testable in isolation (defaults →
      global → repo → flags)
- [ ] Home and repo paths injected via function arguments — no global
      `os.UserHomeDir()` lookups inside the package
- [ ] Doc comments on every exported identifier and field; package comment
      names the precedence rule and the JSON-only contract
- [ ] `go test ./internal/config/... -count=1 -race`, `go vet`,
      `gofmt -l internal/config/` all clean

Commit each box as it flips.

## Tests to write first (red)

Fixtures live in `internal/config/testdata/`. Use `t.TempDir()` for fake
home and repo roots; never read the real `~`.

| Test | Asserts |
|------|---------|
| `TestLoad_NoFiles_ReturnsBuiltinDefaults` | `default_scope = "worktree"`, timeouts populated, `policy.missing_tools = "skip"` |
| `TestLoad_GlobalConfig_OverridesDefaults` | values from `<home>/.config/agent-scan/config.json` win |
| `TestLoad_RepoConfig_OverridesGlobal` | values from `<repo>/.agent-scan.json` win |
| `TestLoad_CLIFlags_OverrideRepoConfig` | flag overrides take final precedence |
| `TestLoad_UnknownField_ReturnsError` | error message names the field and source path |
| `TestLoad_InvalidJSON_ReturnsErrorWithFilePath` | error wraps the parse error with the file location |
| `TestLoad_BadEnumValue_Rejected` | e.g. `default_scope = "wat"` → `ErrInvalidScope` |
| `TestLoad_MissingTimeouts_FillsFromDefaults` | partial config doesn't zero out missing fields |
| `TestConfig_RoundTripsThroughJSONMarshal` | re-marshal yields the same `Config` |

Use `*int` / `*string` for optional fields so "not present" is distinct
from "zero". Apply each layer with `if layer.X != nil { cfg.X = *layer.X }`.

## Files touched

- `internal/config/{config.go, config_test.go, doc.go}`
- `internal/config/testdata/*.json`

## Documentation requirements

- Package comment names the precedence rule, the JSON-only contract, and
  the `DisallowUnknownFields` invariant.
- Every `Config` field has a comment matching the README's field reference
  (godoc is the canonical source; README links to it later).
- `Load` doc comment lists all four sources in order and the error
  conditions it can return.
- Sentinel errors are documented and exported so callers can `errors.Is`.

## Acceptance check

```bash
go test ./internal/config/... -count=1 -race
go vet ./internal/config/...
gofmt -l internal/config/                  # empty
```
