package core

import (
	"strings"
	"testing"
)

func TestRenderHelpCard_CustomTabListsCustomCommands(t *testing.T) {
	e := NewEngine("test", &stubAgent{}, []Platform{&stubPlatformEngine{n: "test"}}, "", LangEnglish)
	e.commands.Add("tokens", "Quota usage report", "", "python usage.py", "", "config")
	e.commands.Add("account", "Switch Claude account", "", "python account.py", "", "config")
	e.commands.Add("hidden", "Should not appear", "", "echo hi", "", "config")
	e.SetDisabledCommands([]string{"hidden"})

	card := e.renderHelpCard()
	if got := countCardActionValues(card, "nav:/help "); got != 5 {
		t.Fatalf("help tab count = %d, want 5 (4 built-in + custom)", got)
	}
	btn, ok := findCardAction(card, "nav:/help custom")
	if !ok {
		t.Fatal("expected custom help tab")
	}
	if btn.Text != "Custom Commands" {
		t.Fatalf("custom tab text = %q", btn.Text)
	}

	custom := e.renderHelpGroupCard("custom")
	text := custom.RenderText()
	if !strings.Contains(text, "**/tokens**  Quota usage report") {
		t.Fatalf("custom help text = %q, want tokens with its config description", text)
	}
	if strings.Index(text, "/account") > strings.Index(text, "/tokens") {
		t.Fatalf("custom commands should be sorted by name: %q", text)
	}
	if strings.Contains(text, "/hidden") {
		t.Fatalf("disabled custom command listed: %q", text)
	}
	if _, ok := findCardAction(custom, "cmd:/tokens"); !ok {
		t.Fatal("custom item should dispatch the command when clicked")
	}
}

func TestRenderHelpCard_NoCustomTabWithoutCustomCommands(t *testing.T) {
	e := NewEngine("test", &stubAgent{}, []Platform{&stubPlatformEngine{n: "test"}}, "", LangEnglish)
	if _, ok := findCardAction(e.renderHelpCard(), "nav:/help custom"); ok {
		t.Fatal("custom tab should be hidden when there are no custom commands")
	}
}

func TestCustomHelpText_PlainTextHelp(t *testing.T) {
	e := NewEngine("test", &stubAgent{}, []Platform{&stubPlatformEngine{n: "test"}}, "", LangChinese)
	if got := e.customHelpText(); got != "" {
		t.Fatalf("customHelpText() = %q, want empty without custom commands", got)
	}
	e.commands.Add("tokens", "订阅额度", "", "python usage.py", "", "config")
	got := e.customHelpText()
	if !strings.HasPrefix(got, "**自定义命令**") || !strings.Contains(got, "/tokens — 订阅额度") {
		t.Fatalf("customHelpText() = %q", got)
	}
}
