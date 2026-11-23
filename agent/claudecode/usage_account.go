package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	// The /usage panel prints reset times with IANA zone names, e.g.
	// "Resets 7pm (America/Chicago)". Release binaries are built with
	// -trimpath, so on Windows time.LoadLocation cannot find
	// $GOROOT/lib/time/zoneinfo.zip and every reset time was dropped (shown
	// as "-"). Embed the zone database so parsing works on every platform.
	_ "time/tzdata"

	"github.com/chenhg5/cc-connect/core"
)

// claudeAccountFile returns the path of Claude Code's global state file, which
// holds the signed-in account under "oauthAccount". It lives in
// $CLAUDE_CONFIG_DIR/.claude.json when that is set, else ~/.claude.json.
func claudeAccountFile(envEntries []string) string {
	for _, kv := range envEntries {
		if v, ok := strings.CutPrefix(kv, "CLAUDE_CONFIG_DIR="); ok {
			if dir := strings.TrimSpace(v); dir != "" {
				return filepath.Join(dir, ".claude.json")
			}
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude.json")
}

// fillUsageAccount adds the signed-in account to a usage report. The /usage
// panel itself shows no account, so /usage replied with "Account: -".
// Only non-secret profile fields are read; credentials live elsewhere.
func fillUsageAccount(report *core.UsageReport, path string) {
	if report == nil || path == "" || report.Email != "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var state struct {
		OAuthAccount struct {
			AccountUUID      string `json:"accountUuid"`
			EmailAddress     string `json:"emailAddress"`
			OrgType          string `json:"organizationType"`
			OrgRateLimitTier string `json:"organizationRateLimitTier"`
		} `json:"oauthAccount"`
	}
	if json.Unmarshal(data, &state) != nil {
		return
	}
	acct := state.OAuthAccount
	report.Email = acct.EmailAddress
	if report.AccountID == "" {
		report.AccountID = acct.AccountUUID
	}
	if report.Plan == "" {
		report.Plan = claudePlanName(acct.OrgType, acct.OrgRateLimitTier)
	}
}

// claudePlanName turns oauthAccount plan fields into a short label, e.g.
// organizationType "claude_max" + tier "default_claude_max_5x" -> "Max 5x".
func claudePlanName(orgType, tier string) string {
	t := strings.ToLower(tier)
	switch {
	case strings.Contains(t, "max_20x"):
		return "Max 20x"
	case strings.Contains(t, "max_5x"):
		return "Max 5x"
	}
	switch strings.ToLower(orgType) {
	case "claude_max":
		return "Max"
	case "claude_pro":
		return "Pro"
	case "claude_team":
		return "Team"
	case "claude_enterprise":
		return "Enterprise"
	}
	return ""
}
