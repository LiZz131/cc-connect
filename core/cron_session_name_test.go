package core

import (
	"testing"
	"time"
)

func TestCronSessionName(t *testing.T) {
	at := time.Date(2026, 11, 23, 9, 0, 0, 0, time.Local)
	cases := []struct {
		name string
		job  *CronJob
		want string
	}{
		{"description wins", &CronJob{ID: "a4edc685", Description: "📰 每日订阅简报", Prompt: "long prompt"}, "📰 每日订阅简报 · 11-23 09:00"},
		{"falls back to prompt", &CronJob{ID: "x", Prompt: "Summarize GitHub trending"}, "Summarize GitHub trending · 11-23 09:00"},
		{"nil job", nil, "cron · 11-23 09:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cronSessionName(tc.job, at); got != tc.want {
				t.Fatalf("cronSessionName() = %q, want %q", got, tc.want)
			}
		})
	}
}
