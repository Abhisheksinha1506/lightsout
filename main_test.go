package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
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
	def := options{cmd: runQuit, timeout: defaultTimeout}
	with := func(change func(o *options)) options {
		o := def
		change(&o)
		return o
	}
	tests := []struct {
		args    []string
		want    options
		wantErr bool
	}{
		{args: nil, want: def},
		{args: []string{"Slack", "Google Chrome"}, want: with(func(o *options) { o.apps = []string{"Slack", "Google Chrome"} })},
		{args: []string{"--version"}, want: options{cmd: showVersion}},
		{args: []string{"-v"}, want: options{cmd: showVersion}},
		{args: []string{"Slack", "--version"}, want: options{cmd: showVersion}},
		{args: []string{"--help"}, want: options{cmd: showHelp}},
		{args: []string{"-h"}, want: options{cmd: showHelp}},
		{args: []string{"-n", "Slack"}, want: with(func(o *options) { o.dryRun = true; o.apps = []string{"Slack"} })},
		{args: []string{"--dry-run", "--yes"}, want: with(func(o *options) { o.dryRun = true; o.yes = true })},
		{args: []string{"-y"}, want: with(func(o *options) { o.yes = true })},
		{args: []string{"--timeout", "30"}, want: with(func(o *options) { o.timeout = 30 * time.Second })},
		{args: []string{"--timeout=5", "Mail"}, want: with(func(o *options) { o.timeout = 5 * time.Second; o.apps = []string{"Mail"} })},
		{args: []string{"keep", "Slack", "Zoom"}, want: options{cmd: keepApps, apps: []string{"Slack", "Zoom"}}},
		{args: []string{"unkeep", "Slack"}, want: options{cmd: unkeepApps, apps: []string{"Slack"}}},
		{args: []string{"completion", "zsh"}, want: options{cmd: printCompletion, apps: []string{"zsh"}}},
		{args: []string{"keep", "-h"}, want: options{cmd: showHelp}},
		{args: []string{"Slack", "keep"}, want: with(func(o *options) { o.apps = []string{"Slack", "keep"} })},
		{args: []string{"--force"}, wantErr: true},
		{args: []string{"Slack", "-x"}, wantErr: true},
		{args: []string{"--timeout"}, wantErr: true},
		{args: []string{"--timeout", "soon"}, wantErr: true},
		{args: []string{"--timeout=0"}, wantErr: true},
		{args: []string{"--timeout=301"}, wantErr: true},
		{args: []string{"keep"}, wantErr: true},
		{args: []string{"keep", "--yes", "Slack"}, wantErr: true},
		{args: []string{"completion"}, wantErr: true},
		{args: []string{"completion", "tcsh"}, wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseArgs(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseArgs(%q) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseArgs(%q) = %+v, want %+v", tt.args, got, tt.want)
		}
	}
}

// The terminal is the listed app that owns the closest ancestor process of
// lightsout.
func TestRunningFrom(t *testing.T) {
	apps := []app{
		{name: "Code", pids: []int{500}},
		{name: "Terminal", pids: []int{300}},
		{name: "Slack", pids: []int{900}},
	}
	tests := []struct {
		ancestors []int
		want      string
	}{
		{ancestors: []int{1000, 999, 300, 1}, want: "Terminal"},
		{ancestors: []int{1000, 600, 500, 300}, want: "Code"},
		{ancestors: []int{1000, 42}, want: ""},
		{ancestors: nil, want: ""},
	}
	for _, tt := range tests {
		if got := runningFrom(apps, tt.ancestors); got != tt.want {
			t.Errorf("runningFrom(%v) = %q, want %q", tt.ancestors, got, tt.want)
		}
	}
}

// fakePlatform records which apps run asked to quit.
type fakePlatform struct {
	apps      []app
	listErr   error
	failQuit  map[string]bool
	quit      []string
	timeout   time.Duration
	afterQuit int
}

func (f *fakePlatform) platform(ancestors []int) platform {
	return platform{
		list: func() ([]app, error) { return f.apps, f.listErr },
		quit: func(name string, timeout time.Duration) error {
			f.quit = append(f.quit, name)
			f.timeout = timeout
			if f.failQuit[name] {
				return errors.New("still open")
			}
			return nil
		},
		afterQuit:     func() { f.afterQuit++ },
		afterQuitNote: "Finder windows",
		ancestors:     ancestors,
	}
}

func TestRun(t *testing.T) {
	running := []app{
		{name: "Mail", pids: []int{10}},
		{name: "Slack", pids: []int{20}},
		{name: "Terminal", pids: []int{30}},
		{name: "Zoom", pids: []int{40}},
	}
	fromTerminal := []int{99, 98, 30}

	tests := []struct {
		name      string
		opts      options
		keep      []string
		ancestors []int
		input     string
		failQuit  map[string]bool
		listErr   error
		wantErr   bool
		wantQuit  []string
		wantAfter int
	}{
		{name: "detected terminal is kept", ancestors: fromTerminal, keep: []string{"Zoom"},
			wantQuit: []string{"Mail", "Slack"}, wantAfter: 1},
		{name: "prompt when not detected", input: "3\n",
			wantQuit: []string{"Mail", "Slack", "Zoom"}, wantAfter: 1},
		{name: "prompt answered none of these", input: "0\n", keep: []string{"Terminal"},
			wantQuit: []string{"Mail", "Slack", "Zoom"}, wantAfter: 1},
		{name: "prompt cancelled", input: "", wantErr: true},
		{name: "--yes refuses to guess", opts: options{yes: true}, wantErr: true},
		{name: "--yes with detected terminal", opts: options{yes: true}, ancestors: fromTerminal,
			wantQuit: []string{"Mail", "Slack", "Zoom"}, wantAfter: 1},
		{name: "--yes when everything else is kept", opts: options{yes: true}, keep: []string{"Mail", "Slack", "Terminal", "Zoom"},
			wantAfter: 1},
		{name: "dry run quits nothing", opts: options{dryRun: true}, ancestors: fromTerminal},
		{name: "dry run without detection does not prompt", opts: options{dryRun: true}},
		{name: "one failure still quits the rest", ancestors: fromTerminal, failQuit: map[string]bool{"Slack": true},
			wantErr: true, wantQuit: []string{"Mail", "Slack", "Zoom"}, wantAfter: 1},
		{name: "listing fails", listErr: errors.New("osascript failed"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakePlatform{apps: running, listErr: tt.listErr, failQuit: tt.failQuit}
			opts := tt.opts
			opts.timeout = 7 * time.Second
			err := run(opts, tt.keep, f.platform(tt.ancestors), strings.NewReader(tt.input), io.Discard)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(f.quit, tt.wantQuit) {
				t.Errorf("quit %q, want %q", f.quit, tt.wantQuit)
			}
			if f.afterQuit != tt.wantAfter {
				t.Errorf("afterQuit ran %d times, want %d", f.afterQuit, tt.wantAfter)
			}
			if len(f.quit) > 0 && f.timeout != 7*time.Second {
				t.Errorf("quit timeout = %v, want 7s", f.timeout)
			}
		})
	}
}

// Running from kitty without a detectable parent (for example inside tmux)
// still keeps kitty open.
func TestRunKeepsKittyFromTERM(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	f := &fakePlatform{apps: []app{{name: "Mail"}, {name: "kitty"}}}
	if err := run(options{yes: true, timeout: time.Second}, nil, f.platform(nil), strings.NewReader(""), io.Discard); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if want := []string{"Mail"}; !reflect.DeepEqual(f.quit, want) {
		t.Errorf("quit %q, want %q", f.quit, want)
	}
}

// Every shell that `lightsout completion` accepts has a script.
func TestCompletionScriptsExist(t *testing.T) {
	for _, shell := range shells {
		script, err := completions.ReadFile(completionFiles[shell])
		if err != nil || !strings.Contains(string(script), "lightsout") {
			t.Errorf("completion script for %s: %v", shell, err)
		}
	}
}

// keep and unkeep edit the config under the user's home folder.
func TestEditKept(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	steps := []options{
		{cmd: keepApps, apps: []string{"Slack"}},
		{cmd: keepApps, apps: []string{"Slack"}}, // already kept
		{cmd: unkeepApps, apps: []string{"Slack"}},
		{cmd: unkeepApps, apps: []string{"Slack"}}, // not kept any more
	}
	for _, opts := range steps {
		if err := editKept(opts); err != nil {
			t.Fatalf("editKept(%+v) error = %v", opts, err)
		}
	}
	contents, err := os.ReadFile(filepath.Join(home, ".config", "lightsout", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), "Slack") {
		t.Errorf("Slack still kept after unkeep:\n%s", contents)
	}
}

// A dry run lists what would be quit and says nothing was.
func TestPrintDryRun(t *testing.T) {
	var out strings.Builder
	printDryRun(&out, []string{"Mail", "Zoom"}, "close all Finder windows")
	for _, want := range []string{"• Mail", "• Zoom", "close all Finder windows.", "Nothing was quit."} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("dry run output missing %q:\n%s", want, out.String())
		}
	}
	out.Reset()
	printDryRun(&out, nil, "")
	if !strings.Contains(out.String(), "no apps would be quit") {
		t.Errorf("empty dry run output:\n%s", out.String())
	}
}
