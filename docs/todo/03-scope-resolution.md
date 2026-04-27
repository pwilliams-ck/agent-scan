# 03 — Scope resolution

> Goal: `internal/scope/` returns the file list for each scope, resolved via
> real `git` plumbing. No mocking of git — tests run against `t.TempDir`
> repos with real `git init`.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [x] All scope tests written; each fails for the expected reason (not a
      compile error)
- [x] All scope tests green, including `-race`
- [x] `internal/scope/` exports `Resolve`, `RepoRoot`, `Scope`, `Parse`,
      `ErrNotARepo`
- [x] Subprocess calls use `exec.CommandContext` and `-z` (NUL-delimited)
      throughout — paths with spaces, newlines, and unicode are correct
- [x] Shared `runGit` helper extracted (refactor); no duplicated argv
      setup
- [x] Doc comments on every exported identifier; package comment in
      `doc.go` names the scope-vs-working-tree v1 rule
- [x] `go test ./internal/scope/... -count=1 -race`, `go vet`,
      `gofmt -l internal/scope/` all clean

Commit each box as it flips.

## Tests to write first (red)

Each subtest creates a fresh `t.TempDir()` repo with `git init` and
populates files via real git operations.

| Test | Asserts |
|------|---------|
| `TestResolve_Staged_ReturnsOnlyIndexed` | only files added with `git add` |
| `TestResolve_Unstaged_ReturnsModifiedNotStaged` | tracked-modified, not in index |
| `TestResolve_Untracked_ReturnsOnlyUntracked` | new files never added |
| `TestResolve_Worktree_IsUnionOfStagedUnstagedUntracked` | union, deduplicated |
| `TestResolve_All_ReturnsAllTrackedFiles` | matches `git ls-files` |
| `TestResolve_FromSubdirectory_ResolvesRepoRoot` | invocation `dir` is a child path |
| `TestResolve_CleanRepo_ReturnsEmptyNotError` | `len(files) == 0`, `err == nil` |
| `TestResolve_NotARepo_ReturnsErrNotARepo` | `errors.Is(err, ErrNotARepo)` |
| `TestResolve_PathsWithSpacesAndUnicode_HandledCorrectly` | NUL split works |
| `TestParseScope_RoundTrip` / `TestParseScope_UnknownStringErrors` | enum sanity |

## Files touched

- `internal/scope/{scope.go, scope_test.go, doc.go}`

## Documentation requirements

- Package comment names the v1 rule from the plan: "scope controls file
  inventory; ecosystem commands run against the working tree."
- `Scope`, `Parse`, `Resolve`, `RepoRoot`, `ErrNotARepo` each have doc
  comments starting with the identifier name.
- One inline comment above the `-z` plumbing explaining why NUL-delimited
  output is required (paths with spaces/newlines).

## Acceptance check

```bash
go test ./internal/scope/... -count=1 -race
go vet ./internal/scope/...
gofmt -l internal/scope/                   # empty
```
