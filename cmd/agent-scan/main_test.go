package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_Version_PrintsAgentScanLine(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"agent-scan", "version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d; want 0", code)
	}
	if !strings.HasPrefix(stdout.String(), "agent-scan ") {
		t.Errorf("stdout = %q; want prefix %q", stdout.String(), "agent-scan ")
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q; want empty", stderr.String())
	}
}

func TestRun_NoArgs_PrintsUsageAndExitsTwo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"agent-scan"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("exit code = %d; want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr.String()), "usage") {
		t.Errorf("stderr = %q; want to contain %q (case-insensitive)", stderr.String(), "usage")
	}
}

func TestRun_UnknownSubcommand_PrintsUsageAndExitsTwo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"agent-scan", "frobnicate"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("exit code = %d; want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr.String()), "usage") {
		t.Errorf("stderr = %q; want to contain %q (case-insensitive)", stderr.String(), "usage")
	}
}
