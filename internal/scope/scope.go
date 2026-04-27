package scope

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// ErrNotARepo is returned by [Resolve] and [RepoRoot] when the supplied
// path is not inside a git working tree.
var ErrNotARepo = errors.New("not a git repository")

// Scope identifies which files an agent-scan invocation should consider.
type Scope int

const (
	// Staged includes files whose index entry differs from HEAD, or every
	// file in the index when the repository has no commits yet.
	Staged Scope = iota
	// Unstaged includes tracked files modified in the working tree but
	// not yet added to the index.
	Unstaged
	// Untracked includes files present in the working tree that git has
	// never seen, excluding paths matched by gitignore rules.
	Untracked
	// Worktree is the union of [Staged], [Unstaged], and [Untracked]. It
	// is the default scope and gives pre-commit confidence over local
	// edits.
	Worktree
	// All includes every tracked file in the repository.
	All
)

// String returns the canonical lowercase name of s, matching the value
// accepted by [Parse] and emitted on the CLI.
func (s Scope) String() string {
	switch s {
	case Staged:
		return "staged"
	case Unstaged:
		return "unstaged"
	case Untracked:
		return "untracked"
	case Worktree:
		return "worktree"
	case All:
		return "all"
	default:
		return fmt.Sprintf("scope(%d)", int(s))
	}
}

// Parse converts the CLI-facing scope string into a [Scope] value. The
// accepted spellings are "staged", "unstaged", "untracked", "worktree",
// and "all"; any other input returns a non-nil error.
func Parse(s string) (Scope, error) {
	switch s {
	case "staged":
		return Staged, nil
	case "unstaged":
		return Unstaged, nil
	case "untracked":
		return Untracked, nil
	case "worktree":
		return Worktree, nil
	case "all":
		return All, nil
	default:
		return 0, fmt.Errorf("unknown scope %q", s)
	}
}

// RepoRoot returns the absolute path to the git working-tree root that
// contains dir. If dir is not inside a git repository, the returned
// error satisfies errors.Is(err, [ErrNotARepo]).
func RepoRoot(dir string) (string, error) {
	out, _, err := runGit(context.Background(), dir, "rev-parse", "--show-toplevel")
	if err != nil {
		if errors.Is(err, ErrNotARepo) {
			return "", err
		}
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// Resolve returns the list of paths in scope s, relative to the git
// working-tree root that contains dir. Paths are returned in sorted
// order with duplicates removed.
//
// dir may be any directory inside the working tree; Resolve walks up to
// the root before invoking git plumbing. The git index and working tree
// are read-only — Resolve never mutates repository state.
func Resolve(dir string, s Scope) ([]string, error) {
	root, err := RepoRoot(dir)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	switch s {
	case Staged:
		return stagedFiles(ctx, root)
	case Unstaged:
		return unstagedFiles(ctx, root)
	case Untracked:
		return untrackedFiles(ctx, root)
	case Worktree:
		staged, err := stagedFiles(ctx, root)
		if err != nil {
			return nil, err
		}
		unstaged, err := unstagedFiles(ctx, root)
		if err != nil {
			return nil, err
		}
		untracked, err := untrackedFiles(ctx, root)
		if err != nil {
			return nil, err
		}
		merged := make([]string, 0, len(staged)+len(unstaged)+len(untracked))
		merged = append(merged, staged...)
		merged = append(merged, unstaged...)
		merged = append(merged, untracked...)
		return dedupSort(merged), nil
	case All:
		files, err := allFiles(ctx, root)
		if err != nil {
			return nil, err
		}
		return dedupSort(files), nil
	default:
		return nil, fmt.Errorf("unknown scope %d", int(s))
	}
}

func stagedFiles(ctx context.Context, root string) ([]string, error) {
	if _, _, err := runGit(ctx, root, "rev-parse", "--verify", "HEAD"); err == nil {
		return runGitNUL(ctx, root, "diff", "--cached", "--name-only", "-z")
	}
	// Pre-initial-commit: no HEAD to diff against, so everything in the
	// index is "staged" by definition.
	return runGitNUL(ctx, root, "ls-files", "--cached", "-z")
}

func unstagedFiles(ctx context.Context, root string) ([]string, error) {
	return runGitNUL(ctx, root, "diff", "--name-only", "-z")
}

func untrackedFiles(ctx context.Context, root string) ([]string, error) {
	return runGitNUL(ctx, root, "ls-files", "--others", "--exclude-standard", "-z")
}

func allFiles(ctx context.Context, root string) ([]string, error) {
	return runGitNUL(ctx, root, "ls-files", "-z")
}

// runGit runs `git` with args inside dir. It returns stdout, stderr, and
// any non-nil error. When git reports that dir is not a working tree,
// the returned error wraps [ErrNotARepo].
func runGit(ctx context.Context, dir string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(msg, "not a git repository") {
			return nil, stderr.Bytes(), ErrNotARepo
		}
		return nil, stderr.Bytes(), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, msg)
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}

// runGitNUL runs git and splits stdout on NUL bytes, dropping empty
// tokens (the trailing one produced by the final NUL).
//
// NUL-delimited output is required: git path arguments may contain
// spaces, tabs, newlines, and arbitrary unicode, none of which can be
// recovered safely from line-delimited output.
func runGitNUL(ctx context.Context, dir string, args ...string) ([]string, error) {
	out, _, err := runGit(ctx, dir, args...)
	if err != nil {
		return nil, err
	}
	return splitNUL(out), nil
}

func splitNUL(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	parts := bytes.Split(b, []byte{0})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		out = append(out, string(p))
	}
	return out
}

func dedupSort(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
