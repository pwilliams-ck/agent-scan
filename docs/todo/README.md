# todo/

Step-by-step build checklist for `agent-scan`. Mirrors the build sequence in
`docs/agent-scan-plan.md` (the design record). One file per step. Work them in
order — later steps depend on earlier ones.

Every step is **TDD**: write the failing test, watch it fail with the expected
error, write the minimum code to pass, refactor with tests green. No
production code without a failing test first.

## Where I am

→ **`03-scope-resolution.md`** — first unchecked box

Update this line whenever a step file's last box is flipped. The next step
file's first unchecked box is always the true "next thing to do".

## How to use

- **Branch first.** Before editing any file in a step, check out a feature branch named for the step — e.g. `git checkout -b feat/scope-resolution` for step 03, `feat/config-loading` for step 04. Never code on `main`. If you forgot and committed on main, stop and move the work onto a branch before continuing.
- **One PR per step.** When a step's last box flips and the commit lands on its branch, push, open a PR, merge into `main`, and delete the branch before starting the next step. The next step branches fresh off `main`, not off the previous step's branch (unless explicitly stacking).
- Open the file named in **Where I am**. Pick up at the first unchecked box.
- Each box is a verifiable end-state — a passing test, a clean lint run, a
  file existing. Not a procedural verb like "run gofmt".
- **Every flipped box must be backed by a commit, but one commit can flip
  multiple coupled boxes.** E.g. "tests green", "binary builds", and
  "vet/gofmt clean" all flip together at the green checkpoint of a TDD
  cycle — that's one commit, not three. Group boxes that are physically the
  same change. Don't publish broken intermediate states; red commits stay
  local. The git log is the audit trail for boxes that *did* flip, not a
  diary of every state change in the working tree.
- When a step file's last box flips, update **Where I am** to point at the
  next file, in the same commit.
- If a step's design changes, update both the step file *and* the plan
  (`docs/agent-scan-plan.md`) — cross out and date, don't delete.

Cross-cutting Go conventions live in the plan's **Coding Standards** section.
Step files reference it; they don't restate it.

## Steps

1. [01-verify-cli-flags.md](01-verify-cli-flags.md) — confirm `claude` and `codex` flags + extension paths
2. [02-scaffold-repo.md](02-scaffold-repo.md) — `go mod init`, `cmd/agent-scan/` stub, plan committed
3. [03-scope-resolution.md](03-scope-resolution.md) — staged/unstaged/untracked/worktree/all
4. [04-config-loading.md](04-config-loading.md) — defaults → global → repo → CLI
5. [05-report-renderers.md](05-report-renderers.md) — JSON v1, Markdown, exit codes
6. [06-quality-profile.md](06-quality-profile.md) — generic + Go/Node/Python/Rust quality checks
7. [07-security-profile.md](07-security-profile.md) — secret scan + ecosystem audits
8. [08-peer-dispatcher.md](08-peer-dispatcher.md) — `peer codex` / `peer claude` + recursion guard
9. [09-skill-templates.md](09-skill-templates.md) — Claude command files + Codex SKILL.md
10. [10-installer.md](10-installer.md) — `install.sh` / `uninstall.sh` with `*.bak` semantics
11. [11-end-to-end.md](11-end-to-end.md) — terminal, Claude, Codex, recursion guard live
