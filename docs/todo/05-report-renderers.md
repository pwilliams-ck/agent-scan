# 05 — Report model and renderers

> Goal: `internal/report/` defines the v1 JSON contract, a Markdown renderer,
> and exit-code mapping. Both renderers are driven from the same `Report`
> struct.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] All renderer tests written; golden file fixtures absent → tests fail
      red as expected
- [ ] All renderer tests green, including `-race`; golden files committed
- [ ] `internal/report/` exports `Report`, `Check`, `Finding`, `SkippedTool`,
      `Format`, `Render`, `ExitCode`
- [ ] JSON output uses `enc.SetEscapeHTML(false)`; empty `findings` /
      `skipped` arrays serialise as `[]`, not `null`
- [ ] Markdown template is a single in-package constant (`text/template`),
      no separate `.tmpl` files
- [ ] Doc comments on every exported identifier; package doc calls this
      out as the **v1 stable contract** (renaming or retyping fields is a
      breaking change)
- [ ] `go test ./internal/report/... -count=1 -race`, `go vet`,
      `gofmt -l internal/report/` all clean

Commit each box as it flips.

## Tests to write first (red)

| Test | Asserts |
|------|---------|
| `TestJSON_GoldenFixture` | render a fixture report; bytes equal `testdata/quality_pass.golden.json` |
| `TestJSON_StableFieldOrder` | re-marshal twice, byte-for-byte equal |
| `TestJSON_EmptyArraysSerialiseAsArrays` | `findings: []`, not `null` |
| `TestMarkdown_RendersAllSections` | header, repo line, result, findings grouped by severity, skipped list |
| `TestMarkdown_NoFindings_StillRendersSummary` | empty case looks right |
| `TestExitCode_PassAndWarning_Zero` | both → 0 |
| `TestExitCode_Fail_One` | → 1 |
| `TestExitCode_Error_Two` | → 2 |
| `TestExitCode_PeerRecursionBlocked_Three` | → 3 |
| `TestSeverityOrder_RendersHighestFirst` | critical → high → medium → low → info |

Support a `-update` flag in tests for golden refreshes:

```go
var update = flag.Bool("update", false, "update golden files")
```

## Files touched

- `internal/report/{report.go, report_test.go, doc.go}`
- `internal/report/testdata/*.golden.json`

## Documentation requirements

- Package comment names this as the **v1 stable contract** — changing field
  names or types is a breaking change.
- Every struct and field has a doc comment. Each JSON tag and its comment
  must agree.
- `Render` doc comment lists every `Format` value and the writer contract
  (caller closes; renderer never closes).
- One inline comment above the encoder explaining why HTML escaping is
  disabled (finding messages contain `<>&` legitimately).

## Acceptance check

```bash
go test ./internal/report/... -count=1 -race
go vet ./internal/report/...
gofmt -l internal/report/                  # empty
```
