# 01 — Verify CLI flags and extension paths

> Goal: confirm the exact installed `claude` and `codex` invocation flags
> before any subprocess code is written, and confirm where each tool reads
> personal extensions from. Fail fast if the draft flags in the plan no
> longer match reality.

This step is investigation, not shipping code. The script we write is the
test — it re-runs in CI later as an install preflight.

## Checklist

- [ ] `scripts/check-cli-flags.sh` exists and exits 0 against current
      `claude --help` and `codex exec --help`
- [ ] `docs/cli-flags.md` records: tool, version, exact flags used by
      `agent-scan peer …`, date checked
- [ ] `~/.claude/commands/` and `~/.codex/skills/` directories confirmed
      readable; paths noted in `cli-flags.md`
- [ ] Plan deltas (if any) applied as dated cross-outs in
      `docs/agent-scan-plan.md` (`~~old~~ new (verified YYYY-MM-DD)`)
- [ ] Files committed

Commit each box as it flips — boxes don't count without a commit.

## What the script asserts

`scripts/check-cli-flags.sh` exits non-zero if:

- `claude --help` is missing any of `-p`, `--no-session-persistence`,
  `--allowedTools`
- `codex exec --help` is missing any of `-C`, `--sandbox`,
  `--ask-for-approval`, `--ephemeral`
- `~/.claude/commands` or `~/.codex/skills` is unreadable

If a flag has been renamed, rename references in the plan with a dated
cross-out — never delete sections.

## Files touched

- `scripts/check-cli-flags.sh` (new)
- `docs/cli-flags.md` (new)
- `docs/agent-scan-plan.md` (deltas only, if any)

## Acceptance check

```bash
bash scripts/check-cli-flags.sh   # exits 0
test -f docs/cli-flags.md
```
