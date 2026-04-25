# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status

Pre-implementation. The repo currently contains only `README.md` and `.gitignore`. The design record lives at `~/.claude/plans/agent-scan.md` and is the source of truth for architecture, CLI surface, profiles, peer-audit behavior, repo layout, build sequence, and acceptance criteria. Per the build sequence (step 2), that plan should be committed into this repo as `agent-scan-plan.md` — read it before starting any non-trivial work, and update it (don't delete sections, cross out and date) when design decisions change.

## Architecture (big picture)

Three layers, one stable core:

1. **CLI core** — a Python stdlib package (`src/agent_scan/`) installed as `~/.local/bin/agent-scan`. Resolves the git repo root, loads optional TOML config, detects stacks (Go/Node/Python/Rust/Docker), runs `quality` or `security` profiles, emits Markdown + JSON. No third-party runtime dependencies.
2. **Skill wrappers** — Claude Code slash commands (`~/.claude/commands/scan-quality.md`, `scan-security.md`, `codex-audit.md`) and a Codex skill (`~/.codex/skills/agent-scan-audit/SKILL.md`). Thin: they shell out to `agent-scan` and defer all defaults/validation to the CLI.
3. **Peer dispatcher** — `agent-scan peer codex …` and `agent-scan peer claude …` spawn the *other* model as a non-interactive read-only subprocess in the same repo. Recursion is blocked via the `AGENT_SCAN_PEER_DEPTH` env var: if it's already set in the parent, the child refuses and exits 3.

Key invariants enforced by the CLI (not by the wrappers):
- Default scope is `worktree` (staged + unstaged + untracked).
- Scope controls file inventory and built-in file checks; ecosystem commands (test runners, linters, audits) run against the working tree because those tools don't operate on the git index.
- Exit codes: `0` pass/warning, `1` blocking findings, `2` runtime error, `3` peer recursion blocked.
- No scan command mutates git state or edits repo files. Peer prompts forbid edit/stage/commit/stash/checkout/reset/merge/rebase/push and forbid recursive peer calls.
- Missing optional tools are reported as `SKIPPED` unless `policy.missing_tools = "fail"`.

JSON output shape is a stable v1 contract — see the plan for the schema. Machine readers parse JSON, never Markdown.

## Development workflow

**TDD is required for every step in the build sequence**: red → green → refactor. No production code without a failing test first. This applies to scope resolution, config loading, report rendering, both profiles, and the peer dispatcher.

Once `pyproject.toml` exists, expected commands (per the plan):

```bash
pytest                              # full suite
pytest tests/test_peer.py::test_X   # single test
ruff check src tests                # lint + docstring (D rule set)
./install.sh                        # install CLI + skills (idempotent, backs up to *.bak)
./uninstall.sh                      # remove and restore *.bak
```

Coverage target: 80% line coverage on `src/agent_scan/`; CI fails below threshold. Shell installers are tested via bats or shell fixtures, not coverage-counted.

Peer subprocess assembly must be verified end-to-end with fake `claude`/`codex` binaries on `PATH` — mocks alone are insufficient (per the plan's test list).

Before wiring peer subprocesses, verify the installed CLI flag syntax with `claude --help` and `codex exec --help` — the draft invocations in the plan are drafts, not confirmed flags.

## Documentation standards (enforced, not aspirational)

- Every public function, class, and module in `src/agent_scan/` carries a docstring. Pick Google or NumPy style and stay consistent. CI enforces via `pydocstyle` or `ruff` `D` rules.
- Inline comments only for non-obvious WHY (constraints, invariants, workarounds). No restating WHAT, no task references like "added for issue #X".
- Each Claude command file and the Codex `SKILL.md` opens its body (after frontmatter) with a single `> Example: …` blockquote.
- No `CHANGELOG.md` or `CONTRIBUTING.md` in v1.

## Non-goals for v1

No cron, no scan history DB, no CI integration package, no Homebrew tap, no Docker image, no third-party Python runtime deps, no automatic fixing, no git mutations, no attempt to attach to an already-open Claude or Codex chat.

## Claude command file gotcha

Claude slash command Markdown supports `$ARGUMENTS`, `$1`, `$2`, etc. It does **not** support shell-only forms like `${@:2}`. Let `agent-scan` validate args and apply defaults; keep the wrappers dumb.
