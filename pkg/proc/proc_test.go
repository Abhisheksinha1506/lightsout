package proc

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

// Ancestors follows parent links from start until there is no parent, the
// lookup fails, or the chain loops.
func TestAncestors(t *testing.T) {
	parents := map[int]int{40: 30, 30: 20, 20: 1, 7: 8, 8: 7}
	parent := func(pid int) (int, error) {
		p, ok := parents[pid]
		if !ok {
			return 0, errors.New("no such process")
		}
		return p, nil
	}

	tests := []struct {
		start int
		want  []int
	}{
		{start: 40, want: []int{40, 30, 20}},
		{start: 99, want: []int{99}},
		{start: 7, want: []int{7, 8}},
	}
	for _, tt := range tests {
		if got := Ancestors(tt.start, parent); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Ancestors(%d) = %v, want %v", tt.start, got, tt.want)
		}
	}
}

// ParentPID must agree with the operating system about this test's own
// parent, on every OS CI runs on.
func TestParentPIDOfThisProcess(t *testing.T) {
	got, err := ParentPID(os.Getpid())
	if err != nil {
		t.Fatalf("ParentPID(self) error = %v", err)
	}
	if want := os.Getppid(); got != want {
		t.Errorf("ParentPID(self) = %d, want %d", got, want)
	}
	chain := Ancestors(os.Getpid(), ParentPID)
	if len(chain) < 2 || chain[1] != os.Getppid() {
		t.Errorf("Ancestors(self) = %v, want it to start with %d, %d", chain, os.Getpid(), os.Getppid())
	}
}
