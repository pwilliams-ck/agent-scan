# agent-scan

## What it is

`agent-scan` is a generic, on-demand pre-commit scanner. It runs from any
directory inside a git repo, resolves the repo root automatically, and reports
quality and security findings as Markdown (for humans) or JSON (for machines).
A single Python stdlib CLI is the stable core; Claude Code and Codex skills are
thin wrappers that call it.

## Why

The headline capability is dual-agent peer audit: Claude can ask Codex to audit
the working tree, and Codex can ask Claude — both through the same CLI, both
read-only, both blocked from recursing back into each other. That makes a
"second-opinion" review part of the normal pre-commit loop instead of a
ceremony. Beyond peer audits, `agent-scan` gives every shell, agent, and
editor the same scan surface so findings are reproducible no matter who or
what kicked off the run.

## Install

### CLI

```bash
git clone <repo-url> agent-scan
cd agent-scan
./install.sh
```

`install.sh` installs the CLI to `~/.local/bin/agent-scan`, copies Claude Skill
command files to `~/.claude/commands/`, and copies the Codex Skill to
`~/.codex/skills/agent-scan-audit/`. Existing files are backed up to `*.bak`.
Re-running is idempotent.

Make sure `~/.local/bin` is on your `PATH`.

### Claude Skill

Installed by `install.sh` to `~/.claude/commands/`:

```text
~/.claude/commands/scan-quality.md
~/.claude/commands/scan-security.md
~/.claude/commands/codex-audit.md
```

These are personal, cross-repo Claude Code slash commands.

### Codex Skill

Installed by `install.sh` to `~/.codex/skills/agent-scan-audit/SKILL.md`.

### Uninstall

```bash
./uninstall.sh
```

Removes the CLI and templates and restores `*.bak` files when present. Leaves
user-created config files in place.

## Usage

Default scope is `worktree` (staged + unstaged + untracked). Default output is
Markdown to stdout.

### Quality scan

```bash
$ agent-scan quality
# agent-scan: quality (worktree)
repo: /Users/me/code/my-app
result: pass
checks: 6 ran, 0 skipped
findings: 0
```

### Security scan

```bash
$ agent-scan security
# agent-scan: security (worktree)
repo: /Users/me/code/my-app
result: warning
findings:
  - medium  src/config.py:14  built-in-secrets  Possible AWS access key
```

### Staged-only scope

```bash
$ agent-scan quality --scope staged
```

### JSON output

```bash
$ agent-scan security --json
{"tool":"agent-scan","version":1,"profile":"security","scope":"worktree",
 "result":"warning","findings":[...], ...}
```

### Peer audit: Claude asks Codex

```bash
$ agent-scan peer codex security
# Codex audit (security, worktree)
- high     internal/auth/token.go:88  hardcoded HMAC secret
- medium   cmd/server/main.go:42      missing context cancellation
```

### Peer audit: Codex asks Claude

```bash
$ agent-scan peer claude quality --scope staged
# Claude audit (quality, staged)
- low  README.md:12  trailing whitespace
```

## Integration

### Claude Code

After install, three slash commands are available in any repo:

- `/scan-quality [--scope SCOPE] [--json]` — runs `agent-scan quality`.
- `/scan-security [--scope SCOPE] [--json]` — runs `agent-scan security`.
- `/codex-audit [quality|security] [--scope SCOPE]` — runs
  `agent-scan peer codex …`. Defaults to `security --scope worktree` when no
  arguments are given.

The commands are thin wrappers; all defaults and validation live in the CLI.

### Codex

The `agent-scan-audit` skill triggers on phrases like "scan my changes," "run
a pre-commit scan," "security audit," or "ask Claude for a second opinion."
It runs `agent-scan` directly and uses `agent-scan peer claude …` for the
peer-audit case.

## Configuration

Config is optional. Repos work with built-in defaults.

Precedence (later wins):

```text
built-in defaults < ~/.config/agent-scan/config.toml < <repo>/.agent-scan.toml < CLI flags
```

Example `.agent-scan.toml`:

```toml
version = 1
default_scope = "worktree"

[policy]
# "skip" reports missing tools as skipped checks; "fail" turns them into errors.
missing_tools = "skip"
# Severities (and synthetic categories) that flip the run to a blocking result.
fail_on = ["critical", "high", "command_failed"]

[timeouts]
# Per-profile timeout in seconds.
quality = 300
security = 600
# Total budget for a peer audit subprocess.
peer = 900

[ignore]
# Glob patterns excluded from file inventory and built-in checks.
paths = ["vendor/**", "node_modules/**", "dist/**", "coverage/**"]

[quality]
# Extra commands to run as part of the quality profile.
extra_commands = []

[security]
# Extra commands to run as part of the security profile.
extra_commands = []
# Tool names to skip even if installed (e.g., "trivy").
disabled_tools = []
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0    | Pass or warning only |
| 1    | Blocking findings or failed checks |
| 2    | Scanner/runtime error (bad args, no git repo, etc.) |
| 3    | Peer recursion blocked |

## Troubleshooting

**Tool not found.** Optional ecosystem tools (e.g. `gosec`, `gitleaks`,
`pip-audit`) are reported as `SKIPPED` by default. Install the tool, or set
`policy.missing_tools = "fail"` if you want missing tools to block.

**Peer recursion blocked (exit 3).** The child process saw
`AGENT_SCAN_PEER_DEPTH=1` in its environment. Peer audits are intentionally
one level deep — the audited model cannot call back into the other. Run the
scan locally instead, or invoke the peer command from a non-peer shell.

**No git repo (exit 2).** `agent-scan` resolves the repo root via
`git rev-parse --show-toplevel`. Run inside a repo, or initialize one with
`git init`.

**Scope returns no files.** Clean working tree against the chosen scope. This
is not an error — `agent-scan` exits 0 with an empty `changed_files` list.
Try `--scope all` if you want a full-repo sweep.

**Peer subprocess timed out.** Default peer timeout is 900 seconds. Raise
`timeouts.peer` in `.agent-scan.toml`, narrow the scope, or check whether the
peer model is hung waiting on auth.

## Contributing

Open an issue or PR. Design decisions live in `agent-scan-plan.md` — read it
before proposing larger changes.
