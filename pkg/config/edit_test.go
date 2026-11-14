package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// keep creates the config when there is none, adds only new names, and
// leaves other settings alone; unkeep removes names and reports which.
func TestAddAndRemoveIgnoredApps(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	added, err := AddIgnoredApps([]string{"Slack", "Google Chrome", "Slack"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Slack", "Google Chrome"}; !reflect.DeepEqual(added, want) {
		t.Errorf("added = %q, want %q", added, want)
	}

	file := filepath.Join(home, ".config", "lightsout", "config.yml")
	contents, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	if strings.Contains(string(contents), "ignoredUpdates") {
		t.Errorf("config mentions ignoredUpdates although it was never set:\n%s", contents)
	}

	if added, _ := AddIgnoredApps([]string{"Slack", "Zoom"}); !reflect.DeepEqual(added, []string{"Zoom"}) {
		t.Errorf("second add = %q, want [Zoom]", added)
	}

	removed, err := RemoveIgnoredApps([]string{"Slack", "Mail"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Slack"}; !reflect.DeepEqual(removed, want) {
		t.Errorf("removed = %q, want %q", removed, want)
	}

	var conf Conf
	if !Exists() {
		t.Fatal("Exists() = false after writing the config")
	}
	Read(&conf)
	if want := []string{"Google Chrome", "Zoom"}; !reflect.DeepEqual(conf.IgnoredApps, want) {
		t.Errorf("IgnoredApps = %q, want %q", conf.IgnoredApps, want)
	}
}

// An existing config.yaml (not .yml) is updated in place, and its other
// settings survive.
func TestAddIgnoredAppsKeepsExistingFileAndSettings(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	dir := filepath.Join(home, ".config", "lightsout")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(file, []byte("ignoredUpdates: true\nignoredApps:\n  - Mail\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := AddIgnoredApps([]string{"Slack"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.yml")); !os.IsNotExist(err) {
		t.Error("a second config.yml was created next to config.yaml")
	}
	var conf Conf
	Exists()
	Read(&conf)
	if !conf.IgnoreUpdates || !reflect.DeepEqual(conf.IgnoredApps, []string{"Mail", "Slack"}) {
		t.Errorf("conf = %+v, want ignoredUpdates true and [Mail Slack]", conf)
	}
}

// A config that isn't valid YAML is reported, not overwritten.
func TestAddIgnoredAppsRefusesBrokenConfig(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	dir := filepath.Join(home, ".config", "lightsout")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "config.yml")
	broken := []byte("ignoredApps: [unclosed\n")
	if err := os.WriteFile(file, broken, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := AddIgnoredApps([]string{"Slack"}); err == nil {
		t.Fatal("AddIgnoredApps() on a broken config = nil error, want an error")
	}
	if got, _ := os.ReadFile(file); string(got) != string(broken) {
		t.Errorf("broken config was rewritten:\n%s", got)
	}
}
