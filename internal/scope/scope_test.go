package scope_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/pwilliams-ck/agent-scan/internal/scope"
)

func TestParseScope_RoundTrip(t *testing.T) {
	cases := []scope.Scope{
		scope.Staged,
		scope.Unstaged,
		scope.Untracked,
		scope.Worktree,
		scope.All,
	}
	for _, want := range cases {
		got, err := scope.Parse(want.String())
		if err != nil {
			t.Fatalf("Parse(%q) errored: %v", want.String(), err)
		}
		if got != want {
			t.Fatalf("round trip mismatch: %v -> %v", want, got)
		}
	}
}

func TestParseScope_UnknownStringErrors(t *testing.T) {
	for _, s := range []string{"", "STAGED", "everything", "worktrees", "stagged", "all "} {
		if _, err := scope.Parse(s); err == nil {
			t.Errorf("Parse(%q) succeeded; want error", s)
		}
	}
}

func TestResolve_Staged_ReturnsOnlyIndexed(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "kept.txt", "kept\n")
	gitIn(t, dir, "add", "kept.txt")
	gitIn(t, dir, "commit", "-m", "init")

	writeFile(t, dir, "added.txt", "added\n")
	gitIn(t, dir, "add", "added.txt")

	writeFile(t, dir, "untracked.txt", "untracked\n")

	got, err := scope.Resolve(dir, scope.Staged)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"added.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("staged files = %v; want %v", got, want)
	}
}

func TestResolve_Unstaged_ReturnsModifiedNotStaged(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "tracked.txt", "v1\n")
	gitIn(t, dir, "add", "tracked.txt")
	gitIn(t, dir, "commit", "-m", "init")

	// Modify tracked.txt without staging.
	writeFile(t, dir, "tracked.txt", "v2\n")
	// Add an untracked file that must not show up in unstaged.
	writeFile(t, dir, "new.txt", "fresh\n")

	got, err := scope.Resolve(dir, scope.Unstaged)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"tracked.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unstaged files = %v; want %v", got, want)
	}
}

func TestResolve_Untracked_ReturnsOnlyUntracked(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "tracked.txt", "tracked\n")
	gitIn(t, dir, "add", "tracked.txt")
	gitIn(t, dir, "commit", "-m", "init")

	// Create a gitignored file and a real untracked file.
	writeFile(t, dir, ".gitignore", "ignored.log\n")
	gitIn(t, dir, "add", ".gitignore")
	gitIn(t, dir, "commit", "-m", "ignore")

	writeFile(t, dir, "ignored.log", "noise\n")
	writeFile(t, dir, "fresh.txt", "fresh\n")

	got, err := scope.Resolve(dir, scope.Untracked)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"fresh.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("untracked files = %v; want %v", got, want)
	}
}

func TestResolve_Worktree_IsUnionOfStagedUnstagedUntracked(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "tracked.txt", "v1\n")
	gitIn(t, dir, "add", "tracked.txt")
	gitIn(t, dir, "commit", "-m", "init")

	// Stage a modification, then re-modify after staging so the same path
	// shows up in both staged and unstaged buckets — Worktree must dedupe.
	writeFile(t, dir, "tracked.txt", "v2\n")
	gitIn(t, dir, "add", "tracked.txt")
	writeFile(t, dir, "tracked.txt", "v3\n")

	// Newly staged file.
	writeFile(t, dir, "added.txt", "added\n")
	gitIn(t, dir, "add", "added.txt")

	// Untracked file.
	writeFile(t, dir, "fresh.txt", "fresh\n")

	got, err := scope.Resolve(dir, scope.Worktree)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"added.txt", "fresh.txt", "tracked.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("worktree files = %v; want %v", got, want)
	}
}

func TestResolve_All_ReturnsAllTrackedFiles(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "a.txt", "a\n")
	writeFile(t, dir, "sub/b.txt", "b\n")
	gitIn(t, dir, "add", "a.txt", "sub/b.txt")
	gitIn(t, dir, "commit", "-m", "init")

	// Untracked files must NOT appear in scope.All.
	writeFile(t, dir, "untracked.txt", "no\n")

	got, err := scope.Resolve(dir, scope.All)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := lsFiles(t, dir)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("all files = %v; want %v", got, want)
	}
}

func TestResolve_FromSubdirectory_ResolvesRepoRoot(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "sub/staged.txt", "s\n")
	gitIn(t, dir, "add", "sub/staged.txt")
	writeFile(t, dir, "sub/fresh.txt", "f\n")

	got, err := scope.Resolve(filepath.Join(dir, "sub"), scope.Worktree)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// Paths must be repo-root-relative regardless of where the caller
	// invoked agent-scan from.
	want := []string{"sub/fresh.txt", "sub/staged.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("worktree from subdir = %v; want %v", got, want)
	}
}

func TestResolve_CleanRepo_ReturnsEmptyNotError(t *testing.T) {
	dir := newRepo(t)
	writeFile(t, dir, "a.txt", "a\n")
	gitIn(t, dir, "add", "a.txt")
	gitIn(t, dir, "commit", "-m", "init")

	got, err := scope.Resolve(dir, scope.Worktree)
	if err != nil {
		t.Fatalf("Resolve clean repo: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("clean repo files = %v; want empty", got)
	}
}

func TestResolve_NotARepo_ReturnsErrNotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := scope.Resolve(dir, scope.Worktree)
	if !errors.Is(err, scope.ErrNotARepo) {
		t.Fatalf("Resolve outside repo: err = %v; want ErrNotARepo", err)
	}
}

func TestResolve_PathsWithSpacesAndUnicode_HandledCorrectly(t *testing.T) {
	dir := newRepo(t)
	names := []string{
		"hello world.txt",
		"ünïçødë.txt",
		"with\ttab.txt",
	}
	for _, n := range names {
		writeFile(t, dir, n, "x\n")
	}
	// Add via "." so we don't have to argv-quote each oddball name.
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-m", "init")

	got, err := scope.Resolve(dir, scope.All)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := append([]string(nil), names...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unicode paths = %v; want %v", got, want)
	}
}

// --- helpers ---------------------------------------------------------------

func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "--quiet")
	gitIn(t, dir, "config", "--local", "user.email", "test@example.com")
	gitIn(t, dir, "config", "--local", "user.name", "agent-scan-test")
	gitIn(t, dir, "config", "--local", "commit.gpgsign", "false")
	return dir
}

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_TERMINAL_PROMPT=0",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// lsFiles returns the set of tracked files reported by `git ls-files -z`,
// sorted, so tests can compare against scope.All without re-encoding the
// expected list by hand.
func lsFiles(t *testing.T, dir string) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	var got []string
	for _, p := range splitNULBytes(out) {
		got = append(got, p)
	}
	sort.Strings(got)
	return got
}

func splitNULBytes(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	var out []string
	start := 0
	for i, c := range b {
		if c == 0 {
			if i > start {
				out = append(out, string(b[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(b) {
		out = append(out, string(b[start:]))
	}
	return out
}
