package claudecode

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chenhg5/cc-connect/core"
)

func TestUsageCache_SlowProbeSurvivesShortCallerDeadline(t *testing.T) {
	var c usageCache
	var calls atomic.Int32
	release := make(chan struct{})
	probe := func(ctx context.Context) (*core.UsageReport, error) {
		calls.Add(1)
		select {
		case <-release:
			return &core.UsageReport{Provider: "claudecode"}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// A caller with a tiny deadline (like the reply footer) gives up...
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := c.get(ctx, probe); err == nil {
		t.Fatal("expected short-deadline caller to time out")
	}
	// ...but the probe keeps running; a concurrent caller joins it instead of starting another.
	got := make(chan *core.UsageReport, 1)
	go func() {
		r, _ := c.get(context.Background(), probe)
		got <- r
	}()
	time.Sleep(20 * time.Millisecond)
	close(release)
	select {
	case r := <-got:
		if r == nil || r.Provider != "claudecode" {
			t.Fatalf("report = %+v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("joined caller never got the probe result")
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("probe ran %d times, want 1 (shared in-flight probe)", n)
	}
	// Fresh result is served from cache without probing again.
	if _, err := c.get(context.Background(), probe); err != nil {
		t.Fatalf("cached get: %v", err)
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("probe ran %d times after cache hit, want 1", n)
	}
}

func TestUsageCache_ServesLastGoodReportOnFailure(t *testing.T) {
	c := usageCache{report: &core.UsageReport{Provider: "old"}, at: time.Now().Add(-2 * claudeUsageCacheTTL)}
	r, err := c.get(context.Background(), func(context.Context) (*core.UsageReport, error) {
		return nil, errors.New("probe broke")
	})
	if err != nil || r == nil || r.Provider != "old" {
		t.Fatalf("get() = %+v, %v; want stale report without error", r, err)
	}
}

func TestUsageCache_ErrorWithoutPriorReport(t *testing.T) {
	var c usageCache
	if _, err := c.get(context.Background(), func(context.Context) (*core.UsageReport, error) {
		return nil, errors.New("probe broke")
	}); err == nil {
		t.Fatal("expected error when no report has ever succeeded")
	}
}
