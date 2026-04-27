// Package scope resolves the file inventory for an agent-scan invocation.
//
// agent-scan exposes five scopes — Staged, Unstaged, Untracked, Worktree,
// and All — corresponding to the --scope flag on the CLI. Worktree is
// the default and is the union of staged, unstaged, and untracked files.
//
// V1 rule: scope controls the file inventory and built-in file-content
// checks; ecosystem commands (test runners, linters, package audits) run
// against the working tree because those tools do not operate directly
// on the git index. See docs/agent-scan-plan.md for the full
// specification.
package scope
