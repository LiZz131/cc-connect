package claudecode

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/chenhg5/cc-connect/core"
)

const (
	// claudeUsageCacheTTL is how long a successful /usage probe result is reused.
	// Each probe starts an interactive Claude Code process, so callers that poll
	// (the reply footer asks every ~30s) must not trigger one per call.
	claudeUsageCacheTTL = 60 * time.Second
	// claudeUsageProbeTimeout bounds a single background probe. Cold starts of
	// Claude Code (SessionStart hooks, first-run checks) can take well over the
	// few seconds a caller is willing to wait.
	claudeUsageProbeTimeout = 45 * time.Second
)

// usageCache shares one in-flight /usage probe between concurrent callers and
// keeps the last result for claudeUsageCacheTTL.
//
// The probe runs on its own context: a caller whose deadline expires (the
// reply footer waits only 1.5s, /usage a few seconds) gets an error, but the
// probe keeps running and fills the cache for the next call. Previously the
// caller's short deadline killed the probe every time, so on slow machines it
// never completed.
type usageCache struct {
	mu       sync.Mutex
	report   *core.UsageReport
	at       time.Time
	err      error
	inflight chan struct{}
}

func (c *usageCache) get(ctx context.Context, probe func(context.Context) (*core.UsageReport, error)) (*core.UsageReport, error) {
	c.mu.Lock()
	if c.report != nil && time.Since(c.at) < claudeUsageCacheTTL {
		r := c.report
		c.mu.Unlock()
		return r, nil
	}
	done := c.inflight
	if done == nil {
		done = make(chan struct{})
		c.inflight = done
		go c.refresh(done, probe)
	}
	c.mu.Unlock()

	select {
	case <-done:
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.err != nil {
			if c.report != nil { // keep serving the last good report rather than failing
				return c.report, nil
			}
			return nil, c.err
		}
		return c.report, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("claudecode: usage probe still running in background, try again shortly: %w", ctx.Err())
	}
}

func (c *usageCache) refresh(done chan struct{}, probe func(context.Context) (*core.UsageReport, error)) {
	ctx, cancel := context.WithTimeout(context.Background(), claudeUsageProbeTimeout)
	defer cancel()
	start := time.Now()
	report, err := probe(ctx)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
	if err == nil {
		c.report, c.at = report, time.Now()
		slog.Debug("claudecode: usage probe finished", "elapsed", time.Since(start))
	} else {
		slog.Warn("claudecode: usage probe failed", "elapsed", time.Since(start), "error", err)
	}
	c.inflight = nil
	close(done)
}

// usageScreenTail returns the last n runes of a probe screen for diagnostics.
func usageScreenTail(screen string, n int) string {
	r := []rune(screen)
	if len(r) <= n {
		return screen
	}
	return "…" + string(r[len(r)-n:])
}
