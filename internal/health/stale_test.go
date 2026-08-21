package health

import (
	"testing"
	"time"
)

func TestIsDisconnected(t *testing.T) {
	now := time.Date(2026, time.August, 21, 15, 0, 0, 0, time.UTC)
	if IsDisconnected(now.Add(-14*time.Second), now, 15*time.Second) {
		t.Fatal("agent should still be healthy before stale threshold")
	}
	if !IsDisconnected(now.Add(-16*time.Second), now, 15*time.Second) {
		t.Fatal("agent should be disconnected after stale threshold")
	}
	if !IsDisconnected(time.Time{}, now, 15*time.Second) {
		t.Fatal("agent without heartbeat should be disconnected")
	}
}
