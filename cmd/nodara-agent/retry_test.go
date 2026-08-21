package main

import (
	"testing"
	"time"
)

func TestJitterStaysWithinTwentyPercent(t *testing.T) {
	base := 10 * time.Second
	minimum := 8 * time.Second
	maximum := 12 * time.Second
	for range 200 {
		delay := jitter(base)
		if delay < minimum || delay > maximum {
			t.Fatalf("jitter delay %s outside [%s, %s]", delay, minimum, maximum)
		}
	}
}
