package wait

import (
	"testing"
	"time"
)

func TestUntil(t *testing.T) {
	tests := []struct {
		name       string
		trueAfter  int // done returns true from this call on; 0 means never
		want       bool
		wantCalled int // minimum number of done calls
	}{
		{name: "already done", trueAfter: 1, want: true, wantCalled: 1},
		{name: "done after a few polls", trueAfter: 3, want: true, wantCalled: 3},
		{name: "never done", trueAfter: 0, want: false, wantCalled: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			done := func() bool {
				calls++
				return tt.trueAfter > 0 && calls >= tt.trueAfter
			}

			got := Until(50*time.Millisecond, time.Millisecond, done)

			if got != tt.want {
				t.Errorf("Until() = %v, want %v", got, tt.want)
			}
			if calls < tt.wantCalled {
				t.Errorf("done called %d times, want at least %d", calls, tt.wantCalled)
			}
		})
	}
}

// A done check that never succeeds must not keep Until running much past its
// timeout.
func TestUntilGivesUpAfterTimeout(t *testing.T) {
	start := time.Now()
	Until(30*time.Millisecond, 5*time.Millisecond, func() bool { return false })
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Until took %v, want about 30ms", elapsed)
	}
}
