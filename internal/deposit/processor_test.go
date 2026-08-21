package deposit

import (
	"testing"
	"time"
)

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		count uint32
		want  time.Duration
	}{
		{1, 10 * time.Second},
		{2, 20 * time.Second},
		{3, 40 * time.Second},
		{20, maxRetryDelay},
	}
	for _, test := range tests {
		if got := retryDelay(test.count); got != test.want {
			t.Errorf("retryDelay(%d) = %s, want %s", test.count, got, test.want)
		}
	}
}

func TestInboxProcessorDefaults(t *testing.T) {
	originalLimit := InboxProcessLimit
	originalInterval := InboxIdleInterval
	t.Cleanup(func() {
		InboxProcessLimit = originalLimit
		InboxIdleInterval = originalInterval
	})

	InboxProcessLimit = 0
	InboxIdleInterval = 0
	if got := effectiveInboxProcessLimit(); got != 200 {
		t.Fatalf("effectiveInboxProcessLimit() = %d, want 200", got)
	}
	if got := effectiveInboxIdleInterval(); got != time.Second {
		t.Fatalf("effectiveInboxIdleInterval() = %s, want 1s", got)
	}
}
