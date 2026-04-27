package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrUnknownField is returned when a config layer contains a JSON field
// not present on [Config]. It wraps the underlying decoder error so the
// field name and the source file path are preserved in the message.
var ErrUnknownField = errors.New("unknown config field")

// ErrInvalidScope is returned when default_scope is not one of the five
// recognized scope names ("staged", "unstaged", "untracked",
// "worktree", "all").
var ErrInvalidScope = errors.New("invalid default_scope")

// Config is the fully-resolved agent-scan configuration after all
// precedence layers have been applied.
type Config struct {
	// Version is the config schema version. Only 1 is recognized in v1.
	Version int `json:"version"`
	// DefaultScope is the scope used when --scope is not passed on the
	// command line. One of "staged", "unstaged", "untracked",
	// "worktree", or "all".
	DefaultScope string `json:"default_scope"`
	// Policy controls how missing tools and severity escalation are
	// handled by the runner.
	Policy Policy `json:"policy"`
	// Timeouts caps wall-clock seconds for each profile and for peer
	// subprocesses.
	Timeouts Timeouts `json:"timeouts"`
	// Ignore lists glob patterns excluded from the file inventory.
	Ignore Ignore `json:"ignore"`
	// Quality holds quality-profile-specific overrides.
	Quality Quality `json:"quality"`
	// Security holds security-profile-specific overrides.
	Security Security `json:"security"`
}

// Policy controls runner behavior for unmet preconditions and severity
// escalation.
type Policy struct {
	// MissingTools is "skip" (report SKIPPED) or "fail" (treat as
	// failure) when an optional tool is not on PATH.
	MissingTools string `json:"missing_tools"`
	// FailOn lists the finding categories that escalate the run to a
	// blocking failure (exit code 1).
	FailOn []string `json:"fail_on"`
}

// Timeouts caps wall-clock seconds for each profile and for peer
// subprocesses.
type Timeouts struct {
	// Quality is the cap for `agent-scan quality`.
	Quality int `json:"quality"`
	// Security is the cap for `agent-scan security`.
	Security int `json:"security"`
	// Peer is the cap for `agent-scan peer …` subprocesses.
	Peer int `json:"peer"`
}

// Ignore declares glob patterns excluded from the file inventory.
type Ignore struct {
	// Paths is a list of doublestar glob patterns relative to the repo
	// root.
	Paths []string `json:"paths"`
}

// Quality holds quality-profile-specific overrides.
type Quality struct {
	// ExtraCommands lists shell commands to run in addition to the
	// built-in checks.
	ExtraCommands []string `json:"extra_commands"`
}

// Security holds security-profile-specific overrides.
type Security struct {
	// ExtraCommands lists shell commands to run in addition to the
	// built-in checks.
	ExtraCommands []string `json:"extra_commands"`
	// DisabledTools names tools that should be skipped even when
	// installed.
	DisabledTools []string `json:"disabled_tools"`
}

// Flags carries CLI flag values that override config-file layers. A
// zero value (empty string, zero int) means "flag not provided" and the
// underlying config layer wins.
type Flags struct {
	// Scope overrides default_scope when non-empty.
	Scope string
}

// Defaults returns the built-in default Config that sits at the bottom
// of the precedence chain.
func Defaults() *Config {
	return &Config{
		Version:      1,
		DefaultScope: "worktree",
		Policy: Policy{
			MissingTools: "skip",
			FailOn:       []string{"critical", "high", "command_failed"},
		},
		Timeouts: Timeouts{
			Quality:  300,
			Security: 600,
			Peer:     900,
		},
		Ignore:   Ignore{Paths: []string{}},
		Quality:  Quality{ExtraCommands: []string{}},
		Security: Security{ExtraCommands: []string{}, DisabledTools: []string{}},
	}
}

// Load resolves the effective Config in precedence order: built-in
// defaults → <home>/.config/agent-scan/config.json → <repo>/.agent-scan.json
// → flags. Missing files at any layer are not errors.
//
// Returned errors include the source file path and may wrap
// [ErrUnknownField] (decoder rejected an unknown field) or
// [ErrInvalidScope] (default_scope is not a recognized value).
func Load(home, repo string, flags Flags) (*Config, error) {
	cfg := Defaults()
	if err := mergeFile(cfg, filepath.Join(home, ".config", "agent-scan", "config.json")); err != nil {
		return nil, err
	}
	if err := mergeFile(cfg, filepath.Join(repo, ".agent-scan.json")); err != nil {
		return nil, err
	}
	applyFlags(cfg, flags)
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks cfg for internal consistency. It currently enforces
// that DefaultScope is one of the five recognized scope names. Other
// fields are accepted as-is and validated by the runner that uses them.
func Validate(cfg *Config) error {
	switch cfg.DefaultScope {
	case "staged", "unstaged", "untracked", "worktree", "all":
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidScope, cfg.DefaultScope)
	}
}

// partialConfig mirrors [Config] but uses pointers and nil-able slices
// so a layer can express "field not set" distinct from "field set to
// the zero value". Decoding into this shape lets the merge step apply
// only the keys the file actually contained.
type partialConfig struct {
	Version      *int             `json:"version,omitempty"`
	DefaultScope *string          `json:"default_scope,omitempty"`
	Policy       *partialPolicy   `json:"policy,omitempty"`
	Timeouts     *partialTimeouts `json:"timeouts,omitempty"`
	Ignore       *partialIgnore   `json:"ignore,omitempty"`
	Quality      *partialQuality  `json:"quality,omitempty"`
	Security     *partialSecurity `json:"security,omitempty"`
}

type partialPolicy struct {
	MissingTools *string  `json:"missing_tools,omitempty"`
	FailOn       []string `json:"fail_on,omitempty"`
}

type partialTimeouts struct {
	Quality  *int `json:"quality,omitempty"`
	Security *int `json:"security,omitempty"`
	Peer     *int `json:"peer,omitempty"`
}

type partialIgnore struct {
	Paths []string `json:"paths,omitempty"`
}

type partialQuality struct {
	ExtraCommands []string `json:"extra_commands,omitempty"`
}

type partialSecurity struct {
	ExtraCommands []string `json:"extra_commands,omitempty"`
	DisabledTools []string `json:"disabled_tools,omitempty"`
}

func mergeFile(cfg *Config, path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var p partialConfig
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return fmt.Errorf("%s: %w", path, classifyJSONErr(err))
	}
	apply(cfg, &p)
	return nil
}

// classifyJSONErr converts a json decoder error into a sentinel-wrapped
// error when possible, so callers can use errors.Is to react to known
// failure modes without string matching.
func classifyJSONErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "unknown field") {
		return fmt.Errorf("%w: %s", ErrUnknownField, err.Error())
	}
	return err
}

func apply(cfg *Config, p *partialConfig) {
	if p.Version != nil {
		cfg.Version = *p.Version
	}
	if p.DefaultScope != nil {
		cfg.DefaultScope = *p.DefaultScope
	}
	if p.Policy != nil {
		if p.Policy.MissingTools != nil {
			cfg.Policy.MissingTools = *p.Policy.MissingTools
		}
		if p.Policy.FailOn != nil {
			cfg.Policy.FailOn = p.Policy.FailOn
		}
	}
	if p.Timeouts != nil {
		if p.Timeouts.Quality != nil {
			cfg.Timeouts.Quality = *p.Timeouts.Quality
		}
		if p.Timeouts.Security != nil {
			cfg.Timeouts.Security = *p.Timeouts.Security
		}
		if p.Timeouts.Peer != nil {
			cfg.Timeouts.Peer = *p.Timeouts.Peer
		}
	}
	if p.Ignore != nil && p.Ignore.Paths != nil {
		cfg.Ignore.Paths = p.Ignore.Paths
	}
	if p.Quality != nil && p.Quality.ExtraCommands != nil {
		cfg.Quality.ExtraCommands = p.Quality.ExtraCommands
	}
	if p.Security != nil {
		if p.Security.ExtraCommands != nil {
			cfg.Security.ExtraCommands = p.Security.ExtraCommands
		}
		if p.Security.DisabledTools != nil {
			cfg.Security.DisabledTools = p.Security.DisabledTools
		}
	}
}

func applyFlags(cfg *Config, flags Flags) {
	if flags.Scope != "" {
		cfg.DefaultScope = flags.Scope
	}
}
