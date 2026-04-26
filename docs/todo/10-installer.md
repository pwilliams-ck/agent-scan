# 10 — install.sh / uninstall.sh

> Goal: the only entrypoint a user runs to set everything up. Builds the
> binary, copies templates, manages backups. Tested with `bats`.

Cross-cutting Go conventions live in `docs/agent-scan-plan.md` →
**Coding Standards**. Shell follows POSIX `sh` (no bash-specific features).

## Checklist

- [ ] `tests/install/install.bats` cases written and red
- [ ] `tests/install/uninstall.bats` cases written and red
- [ ] `install.sh`, `uninstall.sh`, `install-lib.sh` authored; bats suites
      green against `HOME=$(mktemp -d)` fake homes
- [ ] File header in each `.sh` (≤5 lines) names the script's purpose and
      the contract: idempotent, backs up to `*.bak`, refuses to overwrite
      an existing `*.bak`, never touches user config
- [ ] Backup-or-bail logic has a one-line WHY comment for the
      refuse-overwrite rule (don't lose someone's hand-tuned override)
- [ ] `bats tests/install/` runs clean in CI
- [ ] Manual smoke: `HOME=/tmp/fake ./install.sh && HOME=/tmp/fake ./uninstall.sh`
      leaves an empty `/tmp/fake` (modulo `.config/` if user files were placed)

Commit each box as it flips.

## Tests to write first (red)

`tests/install/install.bats`:

| Case | Asserts |
|------|---------|
| `installs the CLI to ~/.local/bin/agent-scan` | binary present and executable |
| `installs Claude command files` | three `.md` files at `~/.claude/commands/` |
| `installs Codex skill` | `~/.codex/skills/agent-scan-audit/SKILL.md` |
| `is idempotent: second run does not re-create *.bak` | backup file count stable across re-runs |
| `backs up an existing file to *.bak before overwriting` | original content preserved at `*.bak` |
| `refuses to overwrite an existing *.bak and exits non-zero` | hard fail with clear message |
| `errors clearly if go is missing from PATH` | non-zero exit, `go: not found` style message |
| `warns (does not fail) if ~/.local/bin is not on PATH` | exits 0 but prints guidance to stderr |

`tests/install/uninstall.bats`:

| Case | Asserts |
|------|---------|
| `removes the installed CLI` | binary absent |
| `removes installed Claude command files` | three files absent |
| `removes installed Codex skill` | skill absent |
| `restores *.bak when present` | original content back at the canonical path |
| `does not delete user-created ~/.config/agent-scan/config.json` | user config left alone |
| `is idempotent` | second run is a no-op, exits 0 |

Use `bats-core` + `bats-assert` + `bats-support` (vendored as git
submodules under `tests/install/helpers/`).

## Files touched

- `install.sh`, `uninstall.sh`, `install-lib.sh`
- `tests/install/{install.bats, uninstall.bats}`
- `tests/install/helpers/` — vendored bats libraries

## Documentation requirements

- File header in each `.sh` states purpose + contract.
- One inline WHY comment above the backup-or-bail logic explaining the
  refuse-overwrite rule.
- README's Install section already covers the user-facing surface; this
  step keeps it accurate. Update README only if behaviour diverges.

## Acceptance check

```bash
bats tests/install/
HOME=/tmp/fake ./install.sh   && test -x /tmp/fake/.local/bin/agent-scan
HOME=/tmp/fake ./uninstall.sh && ! test -e /tmp/fake/.local/bin/agent-scan
```
