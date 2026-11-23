package claudecode

import (
	"testing"
	"time"
)

// Simulates a slow host where Claude Code ignores the first Enters while the
// slash-command menu is opening: the probe must keep pressing Enter (bounded)
// and never re-type "/usage".
func TestNextClaudeUsageProbeAction_RetriesEnterUntilPanel(t *testing.T) {
	state := &claudeUsageProbeState{sentWake: true}
	menu := "> /usage\n> /usage Show session cost, plan usage, and activity stats\n/usage-credits Configure usage credits"
	start := time.Unix(1_800_000_000, 0)

	var typed, enters int
	for ms := 0; ms <= 60_000; ms += 100 {
		switch nextClaudeUsageProbeAction(menu, state, start.Add(time.Duration(ms)*time.Millisecond)) {
		case "/usage":
			typed++
		case "\r":
			enters++
		case "":
		default:
			t.Fatalf("unexpected action at %dms", ms)
		}
	}
	if typed != 1 {
		t.Fatalf("typed /usage %d times, want exactly 1", typed)
	}
	if want := 1 + claudeUsageMaxEnterRetries; enters != want {
		t.Fatalf("pressed Enter %d times, want %d (submit + bounded retries)", enters, want)
	}

	// Once the panel is up, no more input is sent.
	panel := "Current session\n██ 12% used\nResets 7pm (America/Chicago)\nCurrent week (all models)\n12% used"
	fresh := &claudeUsageProbeState{sentWake: true, sentUsage: true, sentSubmit: true}
	if got := nextClaudeUsageProbeAction(panel, fresh, start.Add(time.Hour)); got != "" {
		t.Fatalf("action with panel visible = %q, want none", got)
	}
}
