package version

import (
	"strings"
	"testing"
)

func TestString_NotEmpty(t *testing.T) {
	got := String()
	if got == "" {
		t.Fatal(`String() = ""; want non-empty`)
	}
	if got != "dev" && !strings.HasPrefix(got, "v") {
		t.Errorf("String() = %q; want %q or a value starting with %q", got, "dev", "v")
	}
}
