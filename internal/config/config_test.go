package config_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pwilliams-ck/agent-scan/internal/config"
)

func TestLoad_NoFiles_ReturnsBuiltinDefaults(t *testing.T) {
	cfg, err := config.Load(t.TempDir(), t.TempDir(), config.Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultScope != "worktree" {
		t.Errorf("DefaultScope = %q; want worktree", cfg.DefaultScope)
	}
	if cfg.Policy.MissingTools != "skip" {
		t.Errorf("Policy.MissingTools = %q; want skip", cfg.Policy.MissingTools)
	}
	if cfg.Timeouts.Quality == 0 || cfg.Timeouts.Security == 0 || cfg.Timeouts.Peer == 0 {
		t.Errorf("timeouts not populated: %+v", cfg.Timeouts)
	}
	if cfg.Version != 1 {
		t.Errorf("Version = %d; want 1", cfg.Version)
	}
}

func TestLoad_GlobalConfig_OverridesDefaults(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	writeGlobal(t, home, `{"default_scope":"staged","policy":{"missing_tools":"fail"}}`)

	cfg, err := config.Load(home, repo, config.Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultScope != "staged" {
		t.Errorf("DefaultScope = %q; want staged", cfg.DefaultScope)
	}
	if cfg.Policy.MissingTools != "fail" {
		t.Errorf("Policy.MissingTools = %q; want fail", cfg.Policy.MissingTools)
	}
}

func TestLoad_RepoConfig_OverridesGlobal(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	writeGlobal(t, home, `{"default_scope":"staged"}`)
	writeRepo(t, repo, `{"default_scope":"all"}`)

	cfg, err := config.Load(home, repo, config.Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultScope != "all" {
		t.Errorf("DefaultScope = %q; want all", cfg.DefaultScope)
	}
}

func TestLoad_CLIFlags_OverrideRepoConfig(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	writeRepo(t, repo, `{"default_scope":"all"}`)

	cfg, err := config.Load(home, repo, config.Flags{Scope: "staged"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultScope != "staged" {
		t.Errorf("DefaultScope = %q; want staged", cfg.DefaultScope)
	}
}

func TestLoad_UnknownField_ReturnsError(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	writeRepo(t, repo, `{"defualt_scope":"worktree"}`) // typo: "defualt"

	_, err := config.Load(home, repo, config.Flags{})
	if !errors.Is(err, config.ErrUnknownField) {
		t.Fatalf("err = %v; want ErrUnknownField", err)
	}
	if !strings.Contains(err.Error(), "defualt_scope") {
		t.Errorf("err missing field name: %v", err)
	}
	if !strings.Contains(err.Error(), ".agent-scan.json") {
		t.Errorf("err missing source path: %v", err)
	}
}

func TestLoad_InvalidJSON_ReturnsErrorWithFilePath(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	writeRepo(t, repo, `{not-json`)

	_, err := config.Load(home, repo, config.Flags{})
	if err == nil {
		t.Fatal("Load: nil error; want JSON parse error")
	}
	if !strings.Contains(err.Error(), filepath.Join(repo, ".agent-scan.json")) {
		t.Errorf("err missing file path: %v", err)
	}
}

func TestLoad_BadEnumValue_Rejected(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	writeRepo(t, repo, `{"default_scope":"wat"}`)

	_, err := config.Load(home, repo, config.Flags{})
	if !errors.Is(err, config.ErrInvalidScope) {
		t.Fatalf("err = %v; want ErrInvalidScope", err)
	}
}

func TestLoad_MissingTimeouts_FillsFromDefaults(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	// Only override the quality timeout; security and peer must keep
	// their defaults instead of zeroing out.
	writeRepo(t, repo, `{"timeouts":{"quality":42}}`)

	cfg, err := config.Load(home, repo, config.Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Timeouts.Quality != 42 {
		t.Errorf("Timeouts.Quality = %d; want 42", cfg.Timeouts.Quality)
	}
	if cfg.Timeouts.Security != 600 {
		t.Errorf("Timeouts.Security = %d; want default 600", cfg.Timeouts.Security)
	}
	if cfg.Timeouts.Peer != 900 {
		t.Errorf("Timeouts.Peer = %d; want default 900", cfg.Timeouts.Peer)
	}
}

func TestLoad_AllFields_AppliedFromRepoConfig(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	writeRepo(t, repo, `{
		"version": 1,
		"default_scope": "all",
		"policy": {"missing_tools": "fail", "fail_on": ["critical"]},
		"timeouts": {"quality": 11, "security": 22, "peer": 33},
		"ignore": {"paths": ["vendor/**"]},
		"quality": {"extra_commands": ["go test ./..."]},
		"security": {"extra_commands": ["gosec ./..."], "disabled_tools": ["govulncheck"]}
	}`)

	cfg, err := config.Load(home, repo, config.Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := config.Config{
		Version:      1,
		DefaultScope: "all",
		Policy:       config.Policy{MissingTools: "fail", FailOn: []string{"critical"}},
		Timeouts:     config.Timeouts{Quality: 11, Security: 22, Peer: 33},
		Ignore:       config.Ignore{Paths: []string{"vendor/**"}},
		Quality:      config.Quality{ExtraCommands: []string{"go test ./..."}},
		Security:     config.Security{ExtraCommands: []string{"gosec ./..."}, DisabledTools: []string{"govulncheck"}},
	}
	if !reflect.DeepEqual(*cfg, want) {
		t.Fatalf("merged config mismatch:\n got = %+v\nwant = %+v", *cfg, want)
	}
}

func TestConfig_RoundTripsThroughJSONMarshal(t *testing.T) {
	in := config.Defaults()
	bytes, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out config.Config
	if err := json.Unmarshal(bytes, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(*in, out) {
		t.Fatalf("round trip mismatch:\n in = %+v\nout = %+v", *in, out)
	}
}

// --- helpers ---------------------------------------------------------------

func writeGlobal(t *testing.T, home, body string) {
	t.Helper()
	writeFile(t, filepath.Join(home, ".config", "agent-scan", "config.json"), body)
}

func writeRepo(t *testing.T, repo, body string) {
	t.Helper()
	writeFile(t, filepath.Join(repo, ".agent-scan.json"), body)
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
