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

// Options must never fall through to a real run, where they would be taken
// as app names to keep and every other app would be asked to quit.
func TestParseArgs(t *testing.T) {
	tests := []struct {
		args    []string
		keep    []string
		cmd     command
		wantErr bool
	}{
		{args: nil, keep: nil, cmd: runQuit},
		{args: []string{"Slack", "Google Chrome"}, keep: []string{"Slack", "Google Chrome"}, cmd: runQuit},
		{args: []string{"--version"}, cmd: showVersion},
		{args: []string{"-v"}, cmd: showVersion},
		{args: []string{"Slack", "--version"}, cmd: showVersion},
		{args: []string{"--help"}, cmd: showHelp},
		{args: []string{"-h"}, cmd: showHelp},
		{args: []string{"--force"}, wantErr: true},
		{args: []string{"Slack", "-x"}, wantErr: true},
	}
	for _, tt := range tests {
		keep, cmd, err := parseArgs(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseArgs(%q) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			continue
		}
		if tt.wantErr {
			continue
		}
		if cmd != tt.cmd || !reflect.DeepEqual(keep, tt.keep) {
			t.Errorf("parseArgs(%q) = %q, %v; want %q, %v", tt.args, keep, cmd, tt.keep, tt.cmd)
		}
	}
}
