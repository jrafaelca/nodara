package health

import "time"

func IsDisconnected(lastHeartbeat, now time.Time, staleAfter time.Duration) bool {
	if lastHeartbeat.IsZero() {
		return true
	}
	return now.Sub(lastHeartbeat) > staleAfter
}
