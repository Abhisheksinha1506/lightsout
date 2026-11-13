package config

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// writeConfig writes ~/.config/lightsout/config.yml under home with a single
// ignored app.
func writeConfig(t *testing.T, home, ignoredApp string) {
	t.Helper()
	dir := filepath.Join(home, ".config", "lightsout")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := []byte("ignoredApps:\n  - " + ignoredApp + "\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), contents, 0o644); err != nil {
		t.Fatal(err)
	}
}

// main calls Exists more than once per run; the config must still be found
// and read on every call, otherwise ignoredApps is silently dropped.
func TestRepeatedExistsStillLoadsIgnoredApps(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfig(t, home, "Slack")

	for call := 1; call <= 2; call++ {
		if !Exists() {
			t.Fatalf("Exists() call %d = false, want true", call)
		}
	}

	var conf Conf
	Read(&conf)
	if len(conf.IgnoredApps) != 1 || conf.IgnoredApps[0] != "Slack" {
		t.Fatalf("IgnoredApps = %q, want [\"Slack\"]", conf.IgnoredApps)
	}
}

// A Stat error other than "not exist" must stop lightsout with exit code 1
// instead of panicking or silently skipping the config (which would quit
// ignored apps). Exists calls os.Exit, so it runs in a child copy of this test
// binary.
func TestExistsExitsOnStatError(t *testing.T) {
	if os.Getenv("LIGHTSOUT_CONFIG_TEST_CHILD") == "1" {
		Exists()
		return
	}

	home := t.TempDir()
	// ~/.config/lightsout is a file, so Stat of config.yml inside it fails with ENOTDIR.
	if err := os.MkdirAll(filepath.Join(home, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "lightsout"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestExistsExitsOnStatError$")
	cmd.Env = append(os.Environ(), "LIGHTSOUT_CONFIG_TEST_CHILD=1", "HOME="+home)
	err := cmd.Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("Exists() on Stat error: got %v, want exit status 1", err)
	}
}
