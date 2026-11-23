package claudecode

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chenhg5/cc-connect/core"
)

func TestFillUsageAccount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")
	state := `{"numStartups": 3, "oauthAccount": {"accountUuid": "acc-1", "emailAddress": "me@example.com",
		"organizationType": "claude_max", "organizationRateLimitTier": "default_claude_max_5x"}}`
	if err := os.WriteFile(path, []byte(state), 0o600); err != nil {
		t.Fatal(err)
	}

	r := &core.UsageReport{Provider: "claudecode"}
	fillUsageAccount(r, path)
	if r.Email != "me@example.com" || r.AccountID != "acc-1" || r.Plan != "Max 5x" {
		t.Fatalf("report = %+v", r)
	}

	// Missing or malformed files leave the report untouched.
	r2 := &core.UsageReport{}
	fillUsageAccount(r2, filepath.Join(dir, "missing.json"))
	if r2.Email != "" || r2.Plan != "" {
		t.Fatalf("report from missing file = %+v", r2)
	}
}

func TestClaudeAccountFile_HonorsConfigDir(t *testing.T) {
	got := claudeAccountFile([]string{"FOO=1", "CLAUDE_CONFIG_DIR=/custom/cfg"})
	if want := filepath.Join("/custom/cfg", ".claude.json"); got != want {
		t.Fatalf("claudeAccountFile() = %q, want %q", got, want)
	}
}

func TestClaudePlanName(t *testing.T) {
	cases := map[[2]string]string{
		{"claude_max", "default_claude_max_20x"}: "Max 20x",
		{"claude_max", "default_claude_max_5x"}:  "Max 5x",
		{"claude_max", ""}:                       "Max",
		{"claude_pro", ""}:                       "Pro",
		{"", ""}:                                 "",
	}
	for in, want := range cases {
		if got := claudePlanName(in[0], in[1]); got != want {
			t.Errorf("claudePlanName(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

func TestParseResetTimeWithIANAZone(t *testing.T) {
	// Relies on the embedded tz database (time/tzdata); release binaries are
	// built with -trimpath and cannot read $GOROOT/lib/time/zoneinfo.zip.
	now := time.Date(2026, 11, 22, 18, 10, 0, 0, time.UTC)
	if _, err := parseClaudeUsageResetTime("7pm (America/Chicago)", now); err != nil {
		t.Fatalf("parse reset time: %v", err)
	}
}
