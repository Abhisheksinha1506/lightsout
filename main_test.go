package main

import (
	"errors"
	"reflect"
	"testing"
)

// One app failing to quit must not leave the remaining apps running.
func TestQuitAppsContinuesAfterFailure(t *testing.T) {
	var attempted []string
	quit := func(app string) error {
		attempted = append(attempted, app)
		if app == "Slack" {
			return errors.New("no matching process")
		}
		return nil
	}

	failed := quitApps([]string{"Mail", "Slack", "Zoom"}, quit)

	if want := []string{"Mail", "Slack", "Zoom"}; !reflect.DeepEqual(attempted, want) {
		t.Errorf("attempted = %q, want %q", attempted, want)
	}
	if want := []string{"Slack"}; !reflect.DeepEqual(failed, want) {
		t.Errorf("failed = %q, want %q", failed, want)
	}
}
