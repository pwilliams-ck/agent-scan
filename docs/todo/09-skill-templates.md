# 09 — Claude / Codex skill templates

> Goal: ship the wrappers. They shell out to `agent-scan` and defer all
> defaults and validation to the CLI. Templates are embedded into the
> binary with `//go:embed` so the installer copies from the binary, not
> the working tree.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] Template lint tests written and red
- [ ] Templates authored per the plan; tests green
- [ ] Templates embedded with `//go:embed`; tests verify embed FS contents
      match the on-disk files (no drift)
- [ ] Each template body opens with a `> Example: …` Markdown blockquote
- [ ] No unsupported shell expansion forms (`${@:2}`, `$@`); only
      `$ARGUMENTS`, `$1`, `$2` appear
- [ ] No CLI defaults in wrappers — `$ARGUMENTS` is passed raw to
      `agent-scan`; defaults live in the CLI
- [ ] Doc comments on every exported identifier in `internal/templates`;
      package comment explains the embed-and-copy strategy
- [ ] `go test ./internal/templates/... -count=1`, `go vet`,
      `gofmt -l internal/templates/` all clean

Commit each box as it flips.

## Tests to write first (red)

| Test | Asserts |
|------|---------|
| `TestClaudeCommands_UseDollarArgumentsOrDollarN` | no `${@:` and no bare `$@`; only `$ARGUMENTS` / `$1` / `$2` appear |
| `TestClaudeCommands_OpenWithExampleBlockquote` | first non-empty body line is `> Example: …` |
| `TestClaudeCommands_HaveAllowedToolsFrontmatter` | `allowed-tools:` key present and non-empty |
| `TestCodexSkill_FrontmatterHasNameAndDescription` | both keys present and non-empty |
| `TestTemplates_NoEmoji` | repo style rule: no emoji in templates |
| `TestTemplates_EmbeddedAndOnDiskAgree` | `embed.FS` content equals on-disk file content |

## Files touched

- `templates/claude-skills/{scan-quality.md, scan-security.md, codex-audit.md}`
- `templates/codex-skills/agent-scan-audit/SKILL.md`
- `internal/templates/{templates.go, templates_test.go, doc.go}` — uses
  `//go:embed templates/...`

## Documentation requirements

- Package comment explains why templates are embedded (single-file binary
  distribution) and where the install destinations live
  (`~/.claude/commands/`, `~/.codex/skills/agent-scan-audit/`).
- Comment above each `//go:embed` directive listing what's embedded.
- Each template body opens with `> Example: …` per the plan's Documentation
  Standards section.

## Acceptance check

```bash
go test ./internal/templates/... -count=1
go vet ./internal/templates/...
gofmt -l internal/templates/                # empty
```

End-to-end exercise of the wrappers themselves is deferred to step 11
(after the installer in step 10 wires them up).
