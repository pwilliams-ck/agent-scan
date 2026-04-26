#!/usr/bin/env bash
# check-cli-flags.sh — preflight verification for `agent-scan peer …` subprocess assembly.
#
# Asserts that the installed `claude` and `codex` CLIs still expose the flags
# the peer dispatcher relies on. Re-run this in CI as an install preflight; if
# upstream renames a flag, this script fails before any peer call would.
#
# Exit codes:
#   0  all required flags found and extension dirs readable
#   1  a required flag is missing from `claude --help` or `codex --help` /
#      `codex exec --help`, or an extension dir is unreadable
#   2  a required CLI is not on PATH

set -euo pipefail

fail() {
    echo "check-cli-flags: $*" >&2
    exit 1
}

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "check-cli-flags: required command not on PATH: $1" >&2
        exit 2
    }
}

require_in() {
    # require_in <pattern> <help-output-file> <human-label>
    local pattern="$1" file="$2" label="$3"
    grep -qE -e "$pattern" "$file" || fail "missing flag '$label' in $(basename "$file")"
}

require_dir_readable() {
    local dir="$1"
    [ -d "$dir" ] && [ -r "$dir" ] || fail "extension dir not readable: $dir"
}

require_cmd claude
require_cmd codex

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

claude_help="$tmpdir/claude.help"
codex_help="$tmpdir/codex.help"
codex_exec_help="$tmpdir/codex-exec.help"

# `claude` is shimmed via a shell function in some user shells; bypass it
# with `command` so we always exercise the real binary.
command claude --help >"$claude_help" 2>&1 || fail "claude --help exited non-zero"
codex --help >"$codex_help" 2>&1 || fail "codex --help exited non-zero"
codex exec --help >"$codex_exec_help" 2>&1 || fail "codex exec --help exited non-zero"

# Claude flags consumed by `agent-scan peer claude`.
require_in '(^|[[:space:]])-p,?[[:space:]]|--print( |$)' "$claude_help" "-p / --print"
require_in '--no-session-persistence' "$claude_help" "--no-session-persistence"
require_in '--allowed[Tt]ools|--allowed-tools' "$claude_help" "--allowedTools"

# Codex flags consumed by `agent-scan peer codex`.
# Note: --ask-for-approval is a TOP-LEVEL flag on `codex`, not on `codex exec`.
# Invocation form must place it before the `exec` subcommand:
#   codex --ask-for-approval never exec -C <dir> --sandbox read-only --ephemeral "<prompt>"
require_in '(-a,[[:space:]]+)?--ask-for-approval' "$codex_help" "--ask-for-approval (top-level)"
require_in '(-C,[[:space:]]+)?--cd' "$codex_exec_help" "-C / --cd"
require_in '(-s,[[:space:]]+)?--sandbox' "$codex_exec_help" "--sandbox"
require_in '--ephemeral' "$codex_exec_help" "--ephemeral"

# Extension dirs the installer writes into.
require_dir_readable "$HOME/.claude/commands"
require_dir_readable "$HOME/.codex/skills"

echo "check-cli-flags: ok"
