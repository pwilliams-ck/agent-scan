# 11 — End-to-end verification

> Goal: walk the system through every user-facing entry point with the real
> Claude and Codex binaries. No production code — this is a manual rehearsal
> that catches what unit tests can't.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**.

## Checklist

- [ ] `docs/e2e-checklist.md` exists and records the rehearsal date,
      `claude` version, `codex` version
- [ ] `scripts/e2e.sh` automates the terminal half of the checklist
- [ ] Terminal: every box in the **Terminal** section below passes
- [ ] Claude Code: every box in the **Claude Code** section passes
- [ ] Codex: every box in the **Codex** section passes
- [ ] `git status -s` is byte-identical before and after every scan
      command (read-only invariant)
- [ ] `docs/e2e-checklist.md` committed with the rehearsal date and tool
      versions

Commit each box as it flips.

## Terminal

- [ ] `agent-scan version` prints non-empty version line
- [ ] `agent-scan quality` on a clean repo exits 0, result `pass`
- [ ] Broken `gofmt` in a fixture repo → `agent-scan quality` exits 1 with
      a finding pointing at file:line
- [ ] `agent-scan security --json` produces machine-parseable JSON; pipe
      through `jq -e .` and confirm exit 0
- [ ] `agent-scan quality --scope staged` reports only staged files in
      `changed_files`
- [ ] `AGENT_SCAN_PEER_DEPTH=1 agent-scan peer codex security` exits 3
      with a clear stderr message

## Claude Code

- [ ] `/scan-quality` runs and surfaces findings grouped by severity
- [ ] `/scan-security` runs and leads with high/critical findings
- [ ] `/codex-audit security` actually invokes `codex` and surfaces
      findings under a "Codex audit" section

## Codex

- [ ] Saying "scan my changes" in a Codex session triggers the
      `agent-scan-audit` skill and runs `agent-scan quality --scope worktree`
- [ ] Saying "ask Claude to audit" invokes `agent-scan peer claude security`
      and surfaces a Claude audit
- [ ] Recursion is blocked at runtime: a peer's child cannot call back into
      the other model (verify by reading process tree or by an artificial
      reproduction)

## Files touched

- `docs/e2e-checklist.md` (new)
- `scripts/e2e.sh` (new)

## Documentation requirements

- Each Terminal/Claude/Codex line in `docs/e2e-checklist.md` links to the
  step file (03–10) that owns that behaviour, so any regression here
  points at the right step.
- README's Troubleshooting section is reviewed for accuracy against what
  this rehearsal surfaced — update if reality diverges.

## Acceptance check

Every checkbox above is checked, with the rehearsal date and tool versions
recorded in `docs/e2e-checklist.md`. `git diff` of that file is the audit
trail.

For any failure, file an issue and reopen the responsible step (03–10)
rather than patching here. This step is verification, not new behaviour.
