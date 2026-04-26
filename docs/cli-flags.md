# cli-flags.md

Snapshot of the upstream `claude` and `codex` CLI surfaces that
`agent-scan peer …` depends on. Re-verified by running
`scripts/check-cli-flags.sh`; update this file (and the dated cross-outs
in `docs/agent-scan-plan.md`) whenever a flag changes.

Date checked: 2026-04-26

## Versions

| Tool   | Version reported by `--version`           |
| ------ | ----------------------------------------- |
| claude | `2.1.119 (Claude Code)`                   |
| codex  | `codex-cli 0.125.0`                       |

## Claude flags used by `agent-scan peer claude`

Drafted invocation:

```bash
claude -p --no-session-persistence \
    --allowedTools "Bash(agent-scan:*),Read,Grep,Glob" \
    "<peer prompt>"
```

| Flag                       | Confirmed in `claude --help` | Notes                                                  |
| -------------------------- | ---------------------------- | ------------------------------------------------------ |
| `-p` / `--print`           | yes                          | Non-interactive print-and-exit mode.                   |
| `--no-session-persistence` | yes                          | Only valid with `--print`; suppresses on-disk session. |
| `--allowedTools`           | yes                          | Also accepts `--allowed-tools`.                        |

## Codex flags used by `agent-scan peer codex`

Drafted invocation (corrected — see "Plan delta" below):

```bash
codex --ask-for-approval never exec \
    -C <repo-root> --sandbox read-only --ephemeral \
    "<peer prompt>"
```

| Flag                          | Where it lives           | Confirmed | Notes                                                                                  |
| ----------------------------- | ------------------------ | --------- | -------------------------------------------------------------------------------------- |
| `-a` / `--ask-for-approval`   | **top-level** `codex`    | yes       | `never` skips per-call approvals. Must precede the `exec` subcommand — see plan delta. |
| `-C` / `--cd <DIR>`           | `codex exec`             | yes       | Working root for the agent.                                                            |
| `-s` / `--sandbox <MODE>`     | `codex exec`             | yes       | Use `read-only` for peer audits. Other values: `workspace-write`, `danger-full-access`.|
| `--ephemeral`                 | `codex exec`             | yes       | Skip persisting session files.                                                         |
| `--skip-git-repo-check`       | `codex exec`             | yes (opt) | Not currently used; `agent-scan peer` always runs inside a git repo.                   |

## Plan delta (2026-04-26)

The plan's draft invocation placed `--ask-for-approval never` *after*
`exec`:

```text
codex exec -C <repo-root> --sandbox read-only --ask-for-approval never --ephemeral "<peer prompt>"
```

That form is invalid: `--ask-for-approval` is exposed by `codex --help`
(top-level), not `codex exec --help`. Cross-out applied in
`docs/agent-scan-plan.md` with a dated note pointing at the corrected
form.

## Extension directories

`agent-scan` installs Skill files into the personal extension dirs of
each tool. Both confirmed readable on this machine on 2026-04-26.

| Tool   | Personal extension dir   | Purpose                                                 |
| ------ | ------------------------ | ------------------------------------------------------- |
| claude | `~/.claude/commands/`    | Markdown slash-command files (`/scan-quality`, etc.).   |
| codex  | `~/.codex/skills/`       | Skill packages, each in its own `<skill-name>/` folder. |

## How to re-verify

```bash
bash scripts/check-cli-flags.sh   # exits 0 when surfaces still match
```

If the script fails, read its error, update this file (and the plan)
with the new flag names, and adjust `scripts/check-cli-flags.sh` to
match.
