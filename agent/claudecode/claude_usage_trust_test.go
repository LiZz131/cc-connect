package claudecode

import "testing"

func TestPromptActionForScreenTrustDialog(t *testing.T) {
	cases := []struct {
		name   string
		screen string
		want   string
	}{
		{
			name: "new layout: No selected first",
			screen: "Accessing workspace:\n\nC:\\Temp\\cc-connect-claude-usage-1\n\n" +
				"Quick safety check: Is this a project you created or one you trust?\n\n" +
				"Security guide\n\n❯ No, exit\n  Yes, I trust this folder\n\nEnter to confirm · Esc to cancel",
			want: "\x1b[B\r",
		},
		{
			name: "old layout: Yes selected first",
			screen: "Do you trust the files in this folder?\n\n❯ 1. Yes, proceed\n  2. No, exit\n\n" +
				"Enter to confirm · Esc to exit",
			want: "\r",
		},
		{
			name:   "cursor already on yes",
			screen: "Quick safety check\n\n  No, exit\n❯ Yes, I trust this folder\n\nEnter to confirm",
			want:   "\r",
		},
		{
			name:   "yes above cursor",
			screen: "Quick safety check\n\n  Yes, I trust this folder\n\n❯ No, exit\n\nEnter to confirm",
			want:   "\x1b[A\x1b[A\r",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := promptActionForScreen(tc.screen); got != tc.want {
				t.Fatalf("promptActionForScreen() = %q, want %q", got, tc.want)
			}
		})
	}
}
